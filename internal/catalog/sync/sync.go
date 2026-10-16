// Package sync downloads, verifies, and activates native catalog snapshots.
package sync

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing/fstest"
	"time"

	catalogbundle "polyglot-rosetta/internal/catalog/bundle"
)

const cacheEnvironment = "PROSETTA_CATALOG_CACHE"

var client = &http.Client{
	Timeout: 60 * time.Second,
	CheckRedirect: func(request *http.Request, via []*http.Request) error {
		if len(via) > 0 && (request.URL.Scheme != via[0].URL.Scheme || request.URL.Host != via[0].URL.Host) {
			return fmt.Errorf("catalog redirect changed origin")
		}
		if len(via) >= 10 {
			return fmt.Errorf("too many catalog redirects")
		}
		return nil
	},
}

// DefaultDir returns the cache root shared by catalog and run commands.
func DefaultDir() (string, error) {
	if value := strings.TrimSpace(os.Getenv(cacheEnvironment)); value != "" {
		return filepath.Abs(value)
	}
	root, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "prosetta", "catalog"), nil
}

// ActiveProblemsDir returns the activated snapshot's problems tree, if present.
func ActiveProblemsDir(cacheDir string) (string, bool, error) {
	data, err := os.ReadFile(filepath.Join(cacheDir, "current"))
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	version := strings.TrimSpace(string(data))
	if len(version) != 64 {
		return "", false, fmt.Errorf("invalid activated catalog version")
	}
	if _, err := hex.DecodeString(version); err != nil {
		return "", false, fmt.Errorf("invalid activated catalog version: %w", err)
	}
	root := filepath.Join(cacheDir, "versions", version, "problems")
	if _, err := os.Stat(filepath.Join(root, "info.json")); err != nil {
		return "", false, fmt.Errorf("activated catalog is incomplete: %w", err)
	}
	return root, true, nil
}

// Sync fetches only missing content hashes, validates the complete bundle, and
// activates it after all files are ready.
func Sync(ctx context.Context, baseURL, cacheDir string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("catalog URL must be an HTTP(S) directory URL")
	}
	manifestURL, err := url.JoinPath(baseURL, "manifest.json")
	if err != nil {
		return "", err
	}
	manifestData, err := download(ctx, manifestURL, 1<<20)
	if err != nil {
		return "", fmt.Errorf("fetch catalog manifest: %w", err)
	}
	manifest, err := catalogbundle.ParseManifest(manifestData)
	if err != nil {
		return "", err
	}
	versionHash := sha256.Sum256(manifestData)
	version := hex.EncodeToString(versionHash[:])
	versionDir := filepath.Join(cacheDir, "versions", version)
	if info, err := os.Stat(filepath.Join(versionDir, "problems", "info.json")); err == nil && !info.IsDir() {
		if err := activate(cacheDir, version); err != nil {
			return "", err
		}
		return filepath.Join(versionDir, "problems"), nil
	}

	assets := append([]catalogbundle.Asset{manifest.Index, manifest.Templates}, manifest.Categories...)
	contents := make(map[string][]byte, len(assets))
	for _, asset := range assets {
		data, err := fetchAsset(ctx, baseURL, cacheDir, asset)
		if err != nil {
			return "", err
		}
		contents[asset.Path] = data
	}
	categories := make(map[string][]byte, len(manifest.Categories))
	for _, asset := range manifest.Categories {
		categories[asset.Path] = contents[asset.Path]
	}
	bundle, err := catalogbundle.Load(contents[manifest.Index.Path], contents[manifest.Templates.Path], categories)
	if err != nil {
		return "", fmt.Errorf("validate catalog bundle: %w", err)
	}
	versionsDir := filepath.Join(cacheDir, "versions")
	if err := os.MkdirAll(versionsDir, 0o755); err != nil {
		return "", err
	}
	stage, err := os.MkdirTemp(versionsDir, ".staging-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(stage)
	if err := writeFiles(filepath.Join(stage, "problems"), bundle.Problems); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(stage, "problems", "index.json"), contents[manifest.Index.Path], 0o644); err != nil {
		return "", err
	}
	if err := writeFiles(filepath.Join(stage, "templates"), bundle.Templates); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(stage, "manifest.json"), manifestData, 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(stage, versionDir); err != nil {
		return "", fmt.Errorf("activate catalog snapshot: %w", err)
	}
	if err := activate(cacheDir, version); err != nil {
		return "", err
	}
	return filepath.Join(versionDir, "problems"), nil
}

func fetchAsset(ctx context.Context, baseURL, cacheDir string, asset catalogbundle.Asset) ([]byte, error) {
	blob := filepath.Join(cacheDir, "assets", asset.SHA256)
	if data, err := os.ReadFile(blob); err == nil && validAsset(data, asset) {
		return data, nil
	}
	assetURL, err := url.JoinPath(baseURL, asset.Path)
	if err != nil {
		return nil, err
	}
	data, err := download(ctx, assetURL, asset.Size)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", asset.Path, err)
	}
	if !validAsset(data, asset) {
		return nil, fmt.Errorf("integrity check failed for %s", asset.Path)
	}
	if err := os.MkdirAll(filepath.Dir(blob), 0o755); err != nil {
		return nil, err
	}
	file, err := os.CreateTemp(filepath.Dir(blob), ".asset-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(data); err != nil {
		file.Close()
		return nil, err
	}
	if err := file.Close(); err != nil {
		return nil, err
	}
	if err := os.Chmod(file.Name(), 0o644); err != nil {
		return nil, err
	}
	if err := os.Rename(file.Name(), blob); err != nil {
		return nil, err
	}
	return data, nil
}

func validAsset(data []byte, asset catalogbundle.Asset) bool {
	if int64(len(data)) != asset.Size {
		return false
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]) == asset.SHA256
}

func download(ctx context.Context, location string, limit int64) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", response.StatusCode, location)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("response exceeds size limit")
	}
	return data, nil
}

func writeFiles(root string, files map[string]*fstest.MapFile) error {
	for name, file := range files {
		output := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(output, file.Data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func activate(cacheDir, version string) error {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return err
	}
	file, err := os.CreateTemp(cacheDir, ".current-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.WriteString(version + "\n"); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), filepath.Join(cacheDir, "current"))
}
