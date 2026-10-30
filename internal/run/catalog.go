// Package run coordinates native catalog lookup and solution evaluation.
package run

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"polyglot-rosetta/internal/catalog"
	catalogsync "polyglot-rosetta/internal/catalog/sync"
	"polyglot-rosetta/internal/config"
	"polyglot-rosetta/internal/domain/language"
)

const catalogEnvironment = config.CatalogEnvironment

// ErrCatalogUnavailable means neither a local nor an activated catalog exists.
var ErrCatalogUnavailable = errors.New("no catalog is available")

// Concept identifies a resolved entry in an available catalog.
type Concept struct {
	ID        string
	Reference string
	root      string
	directory string
}

// ResolveConcept locates an ID in the configured, local, or activated catalog.
func ResolveConcept(id string) (Concept, error) {
	roots, explicit, err := catalogRoots()
	if err != nil {
		return Concept{}, err
	}
	for _, root := range roots {
		index, err := catalog.Build(root)
		if err != nil {
			return Concept{}, fmt.Errorf("read catalog %q: %w", root, err)
		}
		entry, err := index.ResolveConcept(id)
		if err == nil {
			fullID, idErr := catalog.ConceptID(entry.Path)
			if idErr != nil {
				return Concept{}, idErr
			}
			return Concept{
				ID: fullID, Reference: entry.Path, root: root,
				directory: filepath.Join(root, filepath.FromSlash(entry.Path)),
			}, nil
		}
		if errors.Is(err, catalog.ErrConceptAmbiguous) {
			return Concept{}, err
		}
		if !errors.Is(err, catalog.ErrConceptNotFound) {
			return Concept{}, err
		}
	}
	if explicit {
		return Concept{}, fmt.Errorf("%w: concept %q was not found under %s", catalog.ErrConceptNotFound, id, catalogEnvironment)
	}
	if len(roots) == 0 {
		return Concept{}, ErrCatalogUnavailable
	}
	return Concept{}, fmt.Errorf("%w: concept %q was not found in the local or activated catalog", catalog.ErrConceptNotFound, id)
}

// LoadSolution reads a catalog reference solution for the selected language.
func LoadSolution(concept Concept, lang language.Language) (string, error) {
	relative := filepath.Join(string(lang), "main."+lang.FileExtension())
	data, err := os.ReadFile(filepath.Join(concept.directory, relative))
	if err != nil {
		return "", fmt.Errorf("read %s catalog solution for %q: %w", lang, concept.Reference, err)
	}
	return string(data), nil
}

func catalogRoots() ([]string, bool, error) {
	if root := strings.TrimSpace(os.Getenv(catalogEnvironment)); root != "" {
		absolute, err := filepath.Abs(root)
		if err != nil {
			return nil, true, fmt.Errorf("resolve %s: %w", catalogEnvironment, err)
		}
		return []string{absolute}, true, nil
	}

	candidates := []string{config.CatalogDir, "./src"}
	if executable, err := os.Executable(); err == nil {
		projectDir := filepath.Dir(filepath.Dir(executable))
		candidates = append(candidates,
			filepath.Join(projectDir, config.CatalogDirName),
			filepath.Join(projectDir, "src"),
		)
	}
	seen := make(map[string]struct{}, len(candidates))
	var roots []string
	for _, candidate := range candidates {
		absolute, err := filepath.Abs(candidate)
		if err != nil {
			return nil, false, fmt.Errorf("resolve catalog candidate %q: %w", candidate, err)
		}
		if _, exists := seen[absolute]; exists {
			continue
		}
		info, err := os.Stat(absolute)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, false, fmt.Errorf("access catalog candidate %q: %w", absolute, err)
		}
		if !info.IsDir() {
			continue
		}
		seen[absolute] = struct{}{}
		roots = append(roots, absolute)
	}
	cacheDir, err := catalogsync.DefaultDir()
	if err != nil && len(roots) == 0 {
		return nil, false, fmt.Errorf("resolve catalog cache: %w", err)
	}
	if err == nil {
		cachedRoot, ready, cacheErr := catalogsync.ActiveProblemsDir(cacheDir)
		if cacheErr != nil && len(roots) == 0 {
			return nil, false, fmt.Errorf("read activated catalog: %w", cacheErr)
		}
		if ready {
			if _, exists := seen[cachedRoot]; !exists {
				roots = append(roots, cachedRoot)
			}
		}
	}
	return roots, false, nil
}
