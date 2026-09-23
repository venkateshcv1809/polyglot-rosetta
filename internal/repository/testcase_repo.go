package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"polyglot-rosetta/internal/domain/testcase"
)

// TestcaseRepository defines the contract for loading test suites.
type TestcaseRepository interface {
	LoadTestCases(conceptDir string) (*testcase.TestSuite, error)
}

// FileResourceRepository implements TestcaseRepository using the local OS filesystem.
type FileResourceRepository struct{}

// NewFileResourceRepository creates a new instance of FileResourceRepository.
func NewFileResourceRepository() *FileResourceRepository {
	return &FileResourceRepository{}
}

// LoadTestCases reads and parses testcases.json from the specified concept directory.
func (r *FileResourceRepository) LoadTestCases(conceptDir string) (*testcase.TestSuite, error) {
	filePath := filepath.Join(conceptDir, "testcases.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read testcases.json at %s: %w", filePath, err)
	}

	var suite testcase.TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse testcases.json: %w", err)
	}

	return &suite, nil
}
