package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConceptScanner_DiscoverConcepts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "content_root_*")
	if err != nil {
		t.Fatalf("failed to temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create nested structure: root/algorithms/two-sum/testcases.json
	conceptPath := filepath.Join(tmpDir, "algorithms", "two-sum")
	if err := os.MkdirAll(conceptPath, 0755); err != nil {
		t.Fatalf("failed to create concept path: %v", err)
	}

	testcasesFile := filepath.Join(conceptPath, "testcases.json")
	if err := os.WriteFile(testcasesFile, []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to write testcases.json: %v", err)
	}

	scanner := NewConceptScanner(tmpDir)
	concepts, err := scanner.DiscoverConcepts()
	if err != nil {
		t.Fatalf("unexpected error during discovery: %v", err)
	}

	if len(concepts) != 1 {
		t.Fatalf("expected 1 discovered concept, got %d", len(concepts))
	}

	if concepts[0] != conceptPath {
		t.Errorf("expected concept path %s, got %s", conceptPath, concepts[0])
	}
}
