package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileResourceRepository_InvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "concept_invalid_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Malformed JSON (missing closing brace)
	invalidJSON := []byte(`{
        "concept": "two-sum",
        "test_cases": [
            {
                "id": "tc_1",
                "description": "basic case"
            }
    }`)

	err = os.WriteFile(filepath.Join(tmpDir, "testcases.json"), invalidJSON, 0644)
	if err != nil {
		t.Fatalf("failed to write testcases.json: %v", err)
	}

	repo := NewFileResourceRepository()
	_, err = repo.LoadTestCases(tmpDir)
	if err == nil {
		t.Errorf("expected decoding error for malformed JSON, got nil")
	}
}
