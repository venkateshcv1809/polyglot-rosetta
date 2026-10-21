package runcmd

import (
	"os"
	"path/filepath"
	"testing"

	"polyglot-rosetta/internal/config"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "prosetta-run-test-*")
	if err != nil {
		panic(err)
	}
	if err := os.Setenv("XDG_CACHE_HOME", dir); err != nil {
		panic(err)
	}
	if err := os.Setenv("PROSETTA_CONFIG", filepath.Join(dir, "config.json")); err != nil {
		panic(err)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func useRunTestCache(t *testing.T) config.CachePaths {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("PROSETTA_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	paths, err := config.NativeCachePaths()
	if err != nil {
		t.Fatal(err)
	}
	return paths
}
