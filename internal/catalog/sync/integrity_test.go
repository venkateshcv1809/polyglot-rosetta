package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	catalogbundle "polyglot-rosetta/internal/catalog/bundle"
	"polyglot-rosetta/internal/config"
	"polyglot-rosetta/internal/testfixture"
)

func TestSyncRechecksAssetsOfActiveSnapshot(t *testing.T) {
	for _, name := range []string{config.CatalogIndexFileName, config.TemplatesArchiveName, "fundamentals" + config.CategoryArchiveSuffix} {
		t.Run(name, func(t *testing.T) {
			output, cache := t.TempDir(), t.TempDir()
			problems, templates := testfixture.Catalog(t)
			if err := catalogbundle.Build(problems, templates, output); err != nil {
				t.Fatal(err)
			}
			fetched := map[string]int{}
			useArchiveTransport(t, output, false, fetched)
			root, err := Sync(context.Background(), "https://static.example/problems/", cache)
			if err != nil {
				t.Fatal(err)
			}
			asset := readTestAsset(t, output, name)
			blob := filepath.Join(cache, config.CatalogAssetsDirName, asset.SHA256)
			original, err := os.ReadFile(blob)
			if err != nil {
				t.Fatal(err)
			}
			corrupt := bytes.Clone(original)
			corrupt[0] ^= 1 // Preserve size so only the digest detects corruption.
			if err := os.WriteFile(blob, corrupt, 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := Sync(context.Background(), "https://static.example/problems/", cache)
			if err != nil || got != root || fetched[name] != 2 {
				t.Fatalf("Sync = %q, %v; fetches = %+v", got, err, fetched)
			}
			repaired, err := os.ReadFile(blob)
			if err != nil || !bytes.Equal(original, repaired) {
				t.Fatalf("corrupt asset was not repaired: %v", err)
			}
		})
	}
}

func TestSyncRejectsCorruptReplacementWithActiveSnapshot(t *testing.T) {
	output, cache := t.TempDir(), t.TempDir()
	problems, templates := testfixture.Catalog(t)
	if err := catalogbundle.Build(problems, templates, output); err != nil {
		t.Fatal(err)
	}
	useArchiveTransport(t, output, false, map[string]int{})
	root, err := Sync(context.Background(), "https://static.example/problems/", cache)
	if err != nil {
		t.Fatal(err)
	}
	asset := readTestAsset(t, output, "fundamentals"+config.CategoryArchiveSuffix)
	if err := os.WriteFile(filepath.Join(cache, config.CatalogAssetsDirName, asset.SHA256), []byte("corrupt"), 0o644); err != nil {
		t.Fatal(err)
	}
	useArchiveTransport(t, output, true, map[string]int{})
	if _, err := Sync(context.Background(), "https://static.example/problems/", cache); err == nil || !strings.Contains(err.Error(), "integrity check failed") {
		t.Fatalf("corrupt replacement error = %v", err)
	}
	active, ready, err := ActiveProblemsDir(cache)
	if err != nil || !ready || active != root {
		t.Fatalf("previous snapshot was not preserved: %q, %t, %v", active, ready, err)
	}
}

func readTestAsset(t *testing.T, output, name string) catalogbundle.Asset {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(output, config.ManifestFileName))
	if err != nil {
		t.Fatal(err)
	}
	var manifest catalogbundle.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, asset := range append([]catalogbundle.Asset{manifest.Index, manifest.Templates}, manifest.Categories...) {
		if asset.Path == name {
			return asset
		}
	}
	t.Fatalf("asset %q missing", name)
	return catalogbundle.Asset{}
}
