package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// ConceptSummary holds metadata for a single concept in the manifest.
type ConceptSummary struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Path     string `json:"path"`
}

// RootManifest represents the compiled index.json structure.
type RootManifest struct {
	Concepts []ConceptSummary `json:"concepts"`
}

// CompileManifest scans the root directory, aggregates concept details, sorts them, and writes index.json.
func CompileManifest(rootDir string, outputPath string) error {
	scanner := NewConceptScanner(rootDir)
	conceptDirs, err := scanner.DiscoverConcepts()
	if err != nil {
		return err
	}

	var summaries []ConceptSummary
	for _, dir := range conceptDirs {
		rel, err := filepath.Rel(rootDir, dir)
		if err != nil {
			rel = dir
		}

		summaries = append(summaries, ConceptSummary{
			ID:       filepath.Base(dir),
			Category: filepath.Base(filepath.Dir(dir)),
			Path:     rel,
		})
	}

	// Sort concepts deterministically by ID for a reproducible manifest structure
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].ID < summaries[j].ID
	})

	manifest := RootManifest{
		Concepts: summaries,
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}
