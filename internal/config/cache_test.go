package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNativeCachePathsDefaultsAndCreatesConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, "cache"))
	t.Setenv(configEnvironment, "")
	paths, err := NativeCachePaths()
	if err != nil {
		t.Fatal(err)
	}
	if paths.Root != filepath.Join(home, "cache", "prosetta") || paths.Catalog != filepath.Join(paths.Root, "catalog") || paths.Toolchain != filepath.Join(paths.Root, "toolchains.json") {
		t.Fatalf("paths = %+v", paths)
	}
	location, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureCacheConfig(); err != nil {
		t.Fatal(err)
	}
	if err := EnsureCacheConfig(); err != nil {
		t.Fatalf("repeat initialization: %v", err)
	}
	data, err := os.ReadFile(location)
	if err != nil {
		t.Fatal(err)
	}
	var settings cacheFile
	if err := json.Unmarshal(data, &settings); err != nil || settings.SchemaVersion != 1 || settings.Cache.Root != paths.Root {
		t.Fatalf("settings = %+v, error = %v", settings, err)
	}
	if info, err := os.Stat(location); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("config permissions = %v, error = %v", info, err)
	}
}

func TestNativeCachePathsReadsOneConfig(t *testing.T) {
	root := t.TempDir()
	location := filepath.Join(t.TempDir(), "config.json")
	t.Setenv(configEnvironment, location)
	writeTestConfig(t, location, root, "catalog-data", "tools/probe.json")
	paths, err := NativeCachePaths()
	if err != nil || paths.Catalog != filepath.Join(root, "catalog-data") || paths.Toolchain != filepath.Join(root, "tools", "probe.json") {
		t.Fatalf("paths = %+v, error = %v", paths, err)
	}
}

func TestNativeCachePathsRejectsUnsafeConfig(t *testing.T) {
	for _, test := range []struct {
		name, catalog, toolchain string
	}{
		{"traversal", "../outside", "toolchains.json"},
		{"absolute", "/elsewhere", "toolchains.json"},
		{"overlap", "catalog", "catalog/toolchains.json"},
	} {
		t.Run(test.name, func(t *testing.T) {
			location := filepath.Join(t.TempDir(), "config.json")
			t.Setenv(configEnvironment, location)
			writeTestConfig(t, location, t.TempDir(), test.catalog, test.toolchain)
			if _, err := NativeCachePaths(); err == nil {
				t.Fatal("invalid cache paths accepted")
			}
		})
	}
}

func TestNativeCachePathsRejectsLegacyEnvironment(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG_CACHE", t.TempDir())
	if _, err := NativeCachePaths(); err == nil || !strings.Contains(err.Error(), "PROSETTA_CATALOG_CACHE") {
		t.Fatalf("legacy setting error = %v", err)
	}
}

func writeTestConfig(t *testing.T, location, root, catalog, toolchain string) {
	t.Helper()
	var settings cacheFile
	settings.SchemaVersion = 1
	settings.Cache.Root = root
	settings.Cache.Catalog = catalog
	settings.Cache.Toolchain = toolchain
	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(location, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
