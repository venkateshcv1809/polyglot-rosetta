package sync

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	catalogbundle "polyglot-rosetta/internal/catalog/bundle"
	"polyglot-rosetta/internal/version"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func useArchiveTransport(t *testing.T, output string, tamper bool, fetched map[string]int) {
	t.Helper()
	previous := client
	client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		name := strings.TrimPrefix(request.URL.Path, "/problems/")
		fetched[name]++
		data, err := os.ReadFile(filepath.Join(output, name))
		if err != nil {
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewReader(nil))}, nil
		}
		if tamper && name == "fundamentals.tar.gz" {
			data = []byte("tampered")
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(data))}, nil
	})}
	t.Cleanup(func() { client = previous })
}

func TestSyncActivatesVerifiedBundleAndReusesAssets(t *testing.T) {
	output := t.TempDir()
	if err := catalogbundle.Build("../../../problems", "../../../templates", output); err != nil {
		t.Fatalf("Build: %v", err)
	}
	fetched := map[string]int{}
	useArchiveTransport(t, output, false, fetched)
	cache := t.TempDir()
	root, err := Sync(context.Background(), "https://static.example/problems/", cache)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "fundamentals", "hello-world", "PROBLEM.md")); err != nil {
		t.Fatalf("problem missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "index.json")); err != nil {
		t.Fatalf("index missing: %v", err)
	}
	active, ok, err := ActiveProblemsDir(cache)
	if err != nil || !ok || active != root {
		t.Fatalf("active = %q, %t, %v", active, ok, err)
	}
	if _, err := Sync(context.Background(), "https://static.example/problems/", cache); err != nil {
		t.Fatalf("second Sync: %v", err)
	}
	if fetched["manifest.json"] != 2 || fetched["fundamentals.tar.gz"] != 1 {
		t.Fatalf("fetch counts = %+v", fetched)
	}
	indexPath := filepath.Join(output, "index.json")
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	indexData = append(indexData, '\n')
	if err := os.WriteFile(indexPath, indexData, 0o644); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(output, "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest catalogbundle.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(indexData)
	manifest.Index.Size = int64(len(indexData))
	manifest.Index.SHA256 = hex.EncodeToString(digest[:])
	manifestData, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, manifestData, 0o644); err != nil {
		t.Fatal(err)
	}
	updatedRoot, err := Sync(context.Background(), "https://static.example/problems/", cache)
	if err != nil || updatedRoot == root {
		t.Fatalf("updated catalog = %q, %v", updatedRoot, err)
	}
	if fetched["index.json"] != 2 || fetched["fundamentals.tar.gz"] != 1 || fetched["templates.tar.gz"] != 1 {
		t.Fatalf("unchanged archives were redownloaded: %+v", fetched)
	}
}

func TestSyncRejectsIncompatibleEngine(t *testing.T) {
	output := t.TempDir()
	if err := catalogbundle.Build("../../../problems", "../../../templates", output); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(output, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest catalogbundle.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.MinEngine = "v99.0.0"
	rewritten, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, rewritten, 0o644); err != nil {
		t.Fatal(err)
	}
	previous := version.Value
	version.Value = "0.1.0"
	t.Cleanup(func() { version.Value = previous })
	useArchiveTransport(t, output, false, map[string]int{})
	cache := t.TempDir()
	if _, err := Sync(context.Background(), "https://static.example/problems/", cache); err == nil || !strings.Contains(err.Error(), "requires Prosetta") {
		t.Fatalf("incompatible engine error = %v", err)
	}
	if _, ok, err := ActiveProblemsDir(cache); ok || err != nil {
		t.Fatalf("unexpected activated catalog: %t, %v", ok, err)
	}
}

func TestSyncRejectsTamperedArchiveWithoutActivating(t *testing.T) {
	output := t.TempDir()
	if err := catalogbundle.Build("../../../problems", "../../../templates", output); err != nil {
		t.Fatal(err)
	}
	useArchiveTransport(t, output, true, map[string]int{})
	cache := t.TempDir()
	if _, err := Sync(context.Background(), "https://static.example/problems/", cache); err == nil {
		t.Fatal("expected archive integrity failure")
	}
	if _, ok, err := ActiveProblemsDir(cache); ok || err != nil {
		t.Fatalf("unexpected activated catalog: %t, %v", ok, err)
	}
}
