package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileResourceRepository_LoadTestCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "concept_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	jsonContent := []byte(`{
		"concept": "two-sum",
		"test_cases": [
			{
				"id": "tc_1",
				"description": "basic case",
				"input": {"x": 1},
				"expected": 2,
				"hidden": false
			}
		]
	}`)

	err = os.WriteFile(filepath.Join(tmpDir, "testcases.json"), jsonContent, 0644)
	if err != nil {
		t.Fatalf("failed to write testcases.json: %v", err)
	}

	repo := NewFileResourceRepository()
	suite, err := repo.LoadTestCases(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error loading test cases: %v", err)
	}

	if suite.Concept != "two-sum" {
		t.Errorf("expected concept 'two-sum', got '%s'", suite.Concept)
	}

	if len(suite.TestCases) != 1 {
		t.Errorf("expected 1 test case, got %d", len(suite.TestCases))
	}
}
