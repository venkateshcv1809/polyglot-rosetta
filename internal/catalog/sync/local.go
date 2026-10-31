package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/config"
)

const catalogEnvironment = config.CatalogEnvironment

// LocalDir returns the authoring problems tree for a development engine.
func LocalDir() (string, error) {
	if root := strings.TrimSpace(os.Getenv(catalogEnvironment)); root != "" {
		absolute, err := filepath.Abs(root)
		if err != nil {
			return "", fmt.Errorf("resolve %s: %w", catalogEnvironment, err)
		}
		return verifiedLocalDir(absolute)
	}
	candidates := []string{config.CatalogDir}
	if executable, err := os.Executable(); err == nil {
		projectDir := filepath.Dir(filepath.Dir(executable))
		candidates = append(candidates, filepath.Join(projectDir, config.CatalogDirName))
	}
	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		absolute, err := filepath.Abs(candidate)
		if err != nil {
			return "", fmt.Errorf("resolve local catalog %q: %w", candidate, err)
		}
		if _, exists := seen[absolute]; exists {
			continue
		}
		seen[absolute] = struct{}{}
		if _, err := os.Stat(filepath.Join(absolute, config.CatalogInfoFileName)); err != nil {
			continue
		}
		return verifiedLocalDir(absolute)
	}
	return "", fmt.Errorf("local catalog not found: run from the repository or set %s", catalogEnvironment)
}

func verifiedLocalDir(root string) (string, error) {
	if _, err := catalog.Build(root); err != nil {
		return "", fmt.Errorf("local catalog %q: %w", root, err)
	}
	return root, nil
}
