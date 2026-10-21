// Package content loads exercise bodies onto canonical catalog metadata.
package content

import (
	"errors"
	"fmt"
	"io/fs"
	"path"

	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/domain/language"
)

// Enrich adds problem text, test cases, solutions, and starters to catalog metadata.
func Enrich(metadata catalog.Catalog, files, templates fs.FS) (catalog.Catalog, error) {
	for id, concept := range metadata.ConceptsByID {
		suite, err := LoadTestSuite(files, concept.Path)
		if err != nil {
			return catalog.Catalog{}, err
		}
		concept.TestCases = suite.TestCases
		concept.Problem, err = readOptional(files, path.Join(concept.Path, "PROBLEM.md"))
		if err != nil {
			return catalog.Catalog{}, err
		}
		concept.Explanation, err = readOptional(files, path.Join(concept.Path, "README.md"))
		if err != nil {
			return catalog.Catalog{}, err
		}
		concept.Solutions = make(map[string]string)
		concept.Starters = make(map[string]string)
		for _, lang := range language.AllSupported() {
			source, err := readOptional(files, path.Join(concept.Path, string(lang), "main."+lang.FileExtension()))
			if err != nil {
				return catalog.Catalog{}, err
			}
			if source != "" {
				concept.Solutions[string(lang)] = source
			}
			starter, present, err := readOptionalPresence(files, path.Join(concept.Path, string(lang), "template."+lang.FileExtension()+".txt"))
			if err != nil {
				return catalog.Catalog{}, err
			}
			if !present && templates != nil {
				starter, present, err = readOptionalPresence(templates, path.Join("concept", string(lang), "template."+lang.FileExtension()))
				if err != nil {
					return catalog.Catalog{}, err
				}
			}
			if present {
				concept.Starters[string(lang)] = starter
			}
		}
		metadata.ConceptsByID[id] = concept
	}
	return metadata, nil
}

func readOptional(files fs.FS, filePath string) (string, error) {
	data, _, err := readOptionalPresence(files, filePath)
	return data, err
}

func readOptionalPresence(files fs.FS, filePath string) (string, bool, error) {
	data, err := fs.ReadFile(files, filePath)
	if err == nil {
		return string(data), true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}
	return "", false, fmt.Errorf("read %q: %w", filePath, err)
}
