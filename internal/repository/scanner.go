package repository

import (
	"io/fs"
	"os"
	"path/filepath"
)

// ConceptScanner handles discovery of concept directories within nested categories.
type ConceptScanner struct {
	RootDir string
}

// NewConceptScanner initializes a scanner with a target root path (e.g., ./content).
func NewConceptScanner(rootDir string) *ConceptScanner {
	return &ConceptScanner{RootDir: rootDir}
}

// DiscoverConcepts recursively walks the directory structure and returns paths
// that contain a target marker file (such as testcases.json).
func (s *ConceptScanner) DiscoverConcepts() ([]string, error) {
	var concepts []string

	err := filepath.WalkDir(s.RootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Gracefully handle permission or missing read errors on individual subtrees
			return nil
		}

		// Look for concept directories identified by the presence of testcases.json
		if !d.IsDir() && d.Name() == "testcases.json" {
			conceptDir := filepath.Dir(path)
			concepts = append(concepts, conceptDir)
		}

		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return concepts, nil
}
