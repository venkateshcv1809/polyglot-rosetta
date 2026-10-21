package run

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/config"
	"polyglot-rosetta/internal/domain/language"
)

func TestResolveConceptAndLoadSolution(t *testing.T) {
	root := writePracticeCatalog(t)
	t.Setenv(catalogEnvironment, root)
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	concept, err := ResolveConcept(id[:10])
	if err != nil {
		t.Fatal(err)
	}
	if concept.ID != id || concept.Reference != "fundamentals/hello-world" || concept.directory != filepath.Join(root, "fundamentals", "hello-world") {
		t.Fatalf("concept = %+v", concept)
	}
	source, err := LoadSolution(concept, language.Go)
	if err != nil || !strings.Contains(source, "Solve") {
		t.Fatalf("solution = %q, error = %v", source, err)
	}
	if _, err := LoadSolution(concept, language.Python); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing solution error = %v", err)
	}
}

func TestResolveConceptWithoutCatalog(t *testing.T) {
	t.Setenv(catalogEnvironment, "")
	useTestNativeCache(t)
	t.Chdir(t.TempDir())
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	if _, err := ResolveConcept(id); !errors.Is(err, ErrCatalogUnavailable) {
		t.Fatalf("error = %v, want ErrCatalogUnavailable", err)
	}
}

func TestResolveConceptUsesActivatedCatalog(t *testing.T) {
	t.Setenv(catalogEnvironment, "")
	t.Chdir(t.TempDir())
	cache := useTestNativeCache(t).Catalog
	version := strings.Repeat("a", 64)
	root := filepath.Join(cache, "versions", version, "problems")
	conceptDir := filepath.Join(root, "fundamentals", "hello-world")
	if err := os.MkdirAll(conceptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCatalogInfo(t, root)
	if err := os.WriteFile(filepath.Join(conceptDir, "testcases.json"), []byte(`[{"id":1,"input":{},"expected":true}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cache, "current"), []byte(version+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	concept, err := ResolveConcept(id)
	if err != nil || concept.directory != conceptDir {
		t.Fatalf("resolved concept = %+v, error = %v", concept, err)
	}
}

func useTestNativeCache(t *testing.T) config.CachePaths {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("PROSETTA_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	paths, err := config.NativeCachePaths()
	if err != nil {
		t.Fatal(err)
	}
	return paths
}

func writePracticeCatalog(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	conceptDir := filepath.Join(root, "fundamentals", "hello-world")
	goDir := filepath.Join(conceptDir, "go")
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(conceptDir, "testcases.json"), []byte(`[{"id":1,"input":{},"expected":true}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(goDir, "main.go"), []byte("package main\ntype Solution struct{}\nfunc (Solution) Solve(any) any { return true }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	writeCatalogInfo(t, root)
	return root
}

func writeCatalogInfo(t *testing.T, root string) {
	t.Helper()
	for name, data := range map[string]string{
		"info.json":                          `{"id":"root","title":"Problems","items":[{"id":"fundamentals","order":1}]}`,
		"fundamentals/info.json":             `{"id":"fundamentals","title":"Fundamentals","items":[{"id":"hello-world","order":1}]}`,
		"fundamentals/hello-world/info.json": `{"id":"hello-world","title":"Hello World","difficulty":"Easy"}`,
	} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

func TestResolveConceptRejectsUnregisteredDirectory(t *testing.T) {
	root := writePracticeCatalog(t)
	dir := filepath.Join(root, "fundamentals", "orphan")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "testcases.json"), []byte(`[{"id":1}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(catalogEnvironment, root)
	id, _ := catalog.ConceptID("fundamentals/orphan")
	if _, err := ResolveConcept(id); !errors.Is(err, catalog.ErrConceptNotFound) {
		t.Fatalf("unregistered concept error = %v", err)
	}
}
