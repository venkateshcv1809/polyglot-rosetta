package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalDirUsesAuthoringTree(t *testing.T) {
	root := writeLocalCatalog(t)
	t.Setenv("PROSETTA_CATALOG", root)
	got, err := LocalDir()
	if err != nil || got != root {
		t.Fatalf("LocalDir = %q, %v", got, err)
	}
}

func TestLocalDirUsesWorkingProblems(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG", "")
	root := writeLocalCatalog(t)
	t.Chdir(filepath.Dir(root))
	got, err := LocalDir()
	if err != nil || got != root {
		t.Fatalf("LocalDir = %q, %v want %q", got, err, root)
	}
}

func TestLocalDirMissing(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG", "")
	t.Chdir(t.TempDir())
	if _, err := LocalDir(); err == nil || !strings.Contains(err.Error(), "local catalog not found") {
		t.Fatalf("missing local catalog error = %v", err)
	}
}

func writeLocalCatalog(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "problems")
	if err := os.MkdirAll(filepath.Join(root, "fundamentals", "hello-world"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"info.json":                          `{"id":"root","title":"Problems","items":[{"id":"fundamentals","order":1}]}`,
		"fundamentals/info.json":             `{"id":"fundamentals","title":"Fundamentals","items":[{"id":"hello-world","order":1}]}`,
		"fundamentals/hello-world/info.json": `{"id":"hello-world","title":"Hello World","difficulty":"Easy"}`,
	}
	for name, data := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
