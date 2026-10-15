package toolchain

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"polyglot-rosetta/internal/domain/language"
)

func TestRequireAvailableProbesOnlyOnCacheMiss(t *testing.T) {
	path := filepath.Join(t.TempDir(), "toolchains.json")
	var lookedUp []string
	lookup := func(binary string) (string, error) {
		lookedUp = append(lookedUp, binary)
		return "/tools/" + binary, nil
	}
	inspect := func(binary, path string) (string, error) { return binary + " 1.0", nil }
	if err := RequireAvailable(path, language.Go, lookup, inspect); err != nil {
		t.Fatalf("first require: %v", err)
	}
	if !slices.Equal(lookedUp, []string{"go", "python3", "rustc", "tsc", "node", "zig"}) {
		t.Fatalf("lookups = %q, want complete probe", lookedUp)
	}
	if err := RequireAvailable(path, language.Go, func(string) (string, error) {
		t.Fatal("cache hit performed lookup")
		return "", nil
	}, func(string, string) (string, error) {
		t.Fatal("cache hit inspected executable")
		return "", nil
	}); err != nil {
		t.Fatalf("cached require: %v", err)
	}
	if err := RequireAvailable(path, language.TypeScript, func(string) (string, error) {
		t.Fatal("complete cache miss for TypeScript")
		return "", nil
	}, func(string, string) (string, error) {
		t.Fatal("complete cache inspected TypeScript")
		return "", nil
	}); err != nil {
		t.Fatalf("cached TypeScript require: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("cache mode = %v, error = %v", info.Mode(), err)
	}
}

func TestCachedUnavailableRequiresManualRefresh(t *testing.T) {
	path := filepath.Join(t.TempDir(), "toolchains.json")
	missing := func(string) (string, error) { return "", errors.New("not found") }
	inspect := func(string, string) (string, error) { t.Fatal("unexpected inspect"); return "", nil }
	if err := RequireAvailable(path, language.Go, missing, inspect); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("first require error = %v", err)
	}
	if err := RequireAvailable(path, language.Go, func(string) (string, error) {
		t.Fatal("cached negative result performed lookup")
		return "", nil
	}, inspect); !errors.Is(err, ErrUnavailable) || !strings.Contains(err.Error(), "prosetta probe --lang=go") {
		t.Fatalf("cached negative error = %v", err)
	}
	if _, err := RefreshCache(path,
		func(binary string) (string, error) { return "/tools/" + binary, nil },
		func(string, string) (string, error) { return "go 1.0", nil },
	); err != nil {
		t.Fatalf("manual refresh: %v", err)
	}
	if err := RequireAvailable(path, language.Go, missing, inspect); err != nil {
		t.Fatalf("require after refresh: %v", err)
	}
}

func TestRequireAvailableRebuildsIncompleteCache(t *testing.T) {
	path := filepath.Join(t.TempDir(), "toolchains.json")
	if err := writeCache(path, cacheSnapshot{
		SchemaVersion: cacheSchema,
		Languages:     map[string]cacheEntry{"go": {Available: true}},
	}); err != nil {
		t.Fatal(err)
	}
	var lookedUp []string
	if err := RequireAvailable(path, language.TypeScript, func(binary string) (string, error) {
		lookedUp = append(lookedUp, binary)
		return "/tools/" + binary, nil
	}, func(string, string) (string, error) { return "1.0", nil }); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(lookedUp, []string{"go", "python3", "rustc", "tsc", "node", "zig"}) {
		t.Fatalf("lookups = %q", lookedUp)
	}
	snapshot, err := readCache(path)
	if err != nil || !snapshot.Languages["go"].Available || !snapshot.Languages["typescript"].Available {
		t.Fatalf("cache = %+v, error = %v", snapshot, err)
	}
}

func TestRequireAvailableRebuildsStaleRequirements(t *testing.T) {
	path := filepath.Join(t.TempDir(), "toolchains.json")
	lookup := func(binary string) (string, error) { return "/tools/" + binary, nil }
	inspect := func(string, string) (string, error) { return "1.0", nil }
	if _, err := RefreshCache(path, lookup, inspect); err != nil {
		t.Fatal(err)
	}
	snapshot, err := readCache(path)
	if err != nil {
		t.Fatal(err)
	}
	entry := snapshot.Languages[string(language.Go)]
	entry.Required = []string{"old-go"}
	snapshot.Languages[string(language.Go)] = entry
	if err := writeCache(path, snapshot); err != nil {
		t.Fatal(err)
	}
	var lookedUp []string
	if err := RequireAvailable(path, language.Go, func(binary string) (string, error) {
		lookedUp = append(lookedUp, binary)
		return "/tools/" + binary, nil
	}, inspect); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(lookedUp, []string{"go", "python3", "rustc", "tsc", "node", "zig"}) {
		t.Fatalf("lookups = %q, want complete probe", lookedUp)
	}
}

func TestRequireAvailableRepairsInvalidCache(t *testing.T) {
	path := filepath.Join(t.TempDir(), "toolchains.json")
	if err := os.WriteFile(path, []byte("not JSON"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := RequireAvailable(path, language.Go,
		func(binary string) (string, error) { return "/tools/" + binary, nil },
		func(string, string) (string, error) { return "go 1.0", nil },
	); err != nil {
		t.Fatalf("repair invalid cache: %v", err)
	}
	if _, err := readCache(path); err != nil {
		t.Fatalf("repaired cache: %v", err)
	}
}

func TestReadCacheRejectsOversizedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "toolchains.json")
	if err := os.WriteFile(path, bytes.Repeat([]byte{' '}, maxCacheBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readCache(path); err == nil || !strings.Contains(err.Error(), "size limit") {
		t.Fatalf("error = %v, want size limit", err)
	}
}

func TestRequireAvailableFailsWhenCacheCannotBeWritten(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(parent, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := RequireAvailable(filepath.Join(parent, "toolchains.json"), language.Go,
		func(binary string) (string, error) { return "/tools/" + binary, nil },
		func(string, string) (string, error) { return "go 1.0", nil },
	)
	if err == nil || !strings.Contains(err.Error(), "save toolchain cache") {
		t.Fatalf("error = %v, want cache write failure", err)
	}
}

func TestDefaultCachePathHonorsOverride(t *testing.T) {
	want := filepath.Join(t.TempDir(), "toolchains.json")
	t.Setenv(cacheEnvironment, want)
	path, err := DefaultCachePath()
	if err != nil || path != want {
		t.Fatalf("path = %q, error = %v; want %q", path, err, want)
	}
}

func TestClearCacheRebuildsOnNextRequire(t *testing.T) {
	path := filepath.Join(t.TempDir(), "toolchains.json")
	lookup := func(binary string) (string, error) { return "/tools/" + binary, nil }
	inspect := func(string, string) (string, error) { return "1.0", nil }
	if _, err := RefreshCache(path, lookup, inspect); err != nil {
		t.Fatal(err)
	}
	if err := ClearCache(path); err != nil {
		t.Fatal(err)
	}
	if err := ClearCache(path); err != nil {
		t.Fatalf("clear absent cache: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stat cleared cache: %v", err)
	}
	lookups := 0
	if err := RequireAvailable(path, language.Go, func(binary string) (string, error) {
		lookups++
		return lookup(binary)
	}, inspect); err != nil {
		t.Fatal(err)
	}
	if lookups != 6 {
		t.Fatalf("lookups = %d, want full probe", lookups)
	}
}

func TestClearCacheRejectsDirectory(t *testing.T) {
	path := t.TempDir()
	if err := ClearCache(path); err == nil || !strings.Contains(err.Error(), "not a file") {
		t.Fatalf("error = %v, want non-file cache path", err)
	}
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		t.Fatalf("cache directory was removed: %v", err)
	}
}
