package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCompileManifest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "manifest_root_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup mock concept directory structure
	conceptPath := filepath.Join(tmpDir, "algorithms", "two-sum")
	if err := os.MkdirAll(conceptPath, 0755); err != nil {
		t.Fatalf("failed to create concept dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(conceptPath, "testcases.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to write testcases.json: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "index.json")
	err = CompileManifest(tmpDir, outputPath)
	if err != nil {
		t.Fatalf("failed to compile manifest: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read generated index.json: %v", err)
	}

	var manifest RootManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("failed to unmarshal index.json: %v", err)
	}

	if len(manifest.Concepts) != 1 {
		t.Fatalf("expected 1 concept in manifest, got %d", len(manifest.Concepts))
	}

	if manifest.Concepts[0].ID != "two-sum" {
		t.Errorf("expected concept ID 'two-sum', got '%s'", manifest.Concepts[0].ID)
	}
}
