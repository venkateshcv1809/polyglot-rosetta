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

// LoadTestCases opens and stream-decodes testcases.json from the specified concept directory.
func (r *FileResourceRepository) LoadTestCases(conceptDir string) (*testcase.TestSuite, error) {
	filePath := filepath.Join(conceptDir, "testcases.json")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open testcases.json at %s: %w", filePath, err)
	}
	defer file.Close()

	var suite testcase.TestSuite
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&suite); err != nil {
		return nil, fmt.Errorf("failed to stream-decode testcases.json syntax error: %w", err)
	}

	return &suite, nil
}
