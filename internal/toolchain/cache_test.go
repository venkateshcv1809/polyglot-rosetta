package toolchain

import (
	"os"
	"path/filepath"
	"testing"

	"polyglot-rosetta/internal/domain/language"
)

func TestCacheStoreOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cache_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cacheFile := filepath.Join(tmpDir, "toolchain_cache.json")

	store, err := NewCacheStore(cacheFile)
	if err != nil {
		t.Fatalf("Failed to initialize cache store: %v", err)
	}

	lang := language.Go
	info := ToolchainInfo{
		Language:  lang,
		Available: true,
		Version:   "go version go1.22.0 linux/amd64",
		Path:      "/usr/local/go/bin/go",
	}

	// Test Set and Get (Zero-latency read)
	store.Set(info)

	got, found := store.Get(lang)
	if !found {
		t.Errorf("Expected to find cached info for %s", lang)
	}
	if got.Version != info.Version {
		t.Errorf("Expected version %s, got %s", info.Version, got.Version)
	}

	// Test Invalidation (Self-healing trigger)
	store.Invalidate(lang)
	_, found = store.Get(lang)
	if found {
		t.Errorf("Expected cache entry to be removed after invalidation")
	}
}

func TestCachePersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cache_persistence_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cacheFile := filepath.Join(tmpDir, "toolchain_cache.json")

	// Initialize first store instance and save an entry
	store1, err := NewCacheStore(cacheFile)
	if err != nil {
		t.Fatalf("Failed to create store 1: %v", err)
	}

	info := ToolchainInfo{
		Language:  language.Python,
		Available: true,
		Version:   "Python 3.10.12",
		Path:      "/usr/bin/python3",
	}
	store1.Set(info)

	// Initialize second store instance pointing to the same file to test reloading from disk
	store2, err := NewCacheStore(cacheFile)
	if err != nil {
		t.Fatalf("Failed to reload cache store from disk: %v", err)
	}

	got, found := store2.Get(language.Python)
	if !found {
		t.Errorf("Expected persisted cache to load successfully from disk")
	}
	if got.Version != info.Version {
		t.Errorf("Expected persisted version %s, got %s", info.Version, got.Version)
	}
}
