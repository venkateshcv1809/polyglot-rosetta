package bundle

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"polyglot-rosetta/internal/catalog"
)

func TestBuildCreatesCategoryAndTemplateArchives(t *testing.T) {
	output := t.TempDir()
	if err := Build("../../../problems", "../../../templates", output); err != nil {
		t.Fatalf("Build: %v", err)
	}
	var manifest Manifest
	data, err := os.ReadFile(filepath.Join(output, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseManifest(data); err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if manifest.SchemaVersion != 1 || manifest.Index.Path != "index.json" || manifest.Templates.Path != "templates.tar.gz" || len(manifest.Categories) != 1 || manifest.Categories[0].Path != "fundamentals.tar.gz" {
		t.Fatalf("manifest = %+v", manifest)
	}
	for _, asset := range append([]Asset{manifest.Index, manifest.Templates}, manifest.Categories...) {
		data, err := os.ReadFile(filepath.Join(output, asset.Path))
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(data)
		if int64(len(data)) != asset.Size || hex.EncodeToString(hash[:]) != asset.SHA256 {
			t.Fatalf("invalid integrity metadata for %s", asset.Path)
		}
	}
	if got := archiveNames(t, filepath.Join(output, "templates.tar.gz")); !contains(got, "concept/go/template.go") {
		t.Fatalf("common Go starter absent: %v", got)
	}
	conceptFiles := archiveNames(t, filepath.Join(output, "fundamentals.tar.gz"))
	if !contains(conceptFiles, "fundamentals/hello-world/PROBLEM.md") || contains(conceptFiles, "fundamentals/hello-world/go/template.go.txt") {
		t.Fatalf("unexpected concept archive contents: %v", conceptFiles)
	}
	indexData, _ := os.ReadFile(filepath.Join(output, manifest.Index.Path))
	templateData, _ := os.ReadFile(filepath.Join(output, manifest.Templates.Path))
	categoryData, _ := os.ReadFile(filepath.Join(output, manifest.Categories[0].Path))
	loaded, err := Load(indexData, templateData, map[string][]byte{manifest.Categories[0].Path: categoryData})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	if loaded.Catalog.ConceptsByID[id].Starters["go"] == "" {
		t.Fatal("common Go starter not loaded from archive")
	}
	var altered catalog.Catalog
	if err := json.Unmarshal(indexData, &altered); err != nil {
		t.Fatal(err)
	}
	concept := altered.ConceptsByID[id]
	concept.Title = "Stale title"
	altered.ConceptsByID[id] = concept
	staleIndex, err := json.Marshal(altered)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Load(staleIndex, templateData, map[string][]byte{manifest.Categories[0].Path: categoryData}); err == nil {
		t.Fatal("expected stale index to be rejected")
	}
	first, err := os.ReadFile(filepath.Join(output, "fundamentals.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if err := Build("../../../problems", "../../../templates", output); err != nil {
		t.Fatalf("second Build: %v", err)
	}
	second, err := os.ReadFile(filepath.Join(output, "fundamentals.tar.gz"))
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("archive is not deterministic: %v", err)
	}
}

func TestWriteArchiveRejectsSymlinks(t *testing.T) {
	source := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(source, "escape")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := writeArchive(t.TempDir(), "source.tar.gz", source, "source"); err == nil {
		t.Fatal("expected symlink rejection")
	}
}

func archiveNames(t *testing.T, filePath string) []string {
	t.Helper()
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	zip, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer zip.Close()
	reader := tar.NewReader(zip)
	var names []string
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, header.Name)
	}
	return names
}

func contains(items []string, item string) bool {
	for _, candidate := range items {
		if candidate == item {
			return true
		}
	}
	return false
}
