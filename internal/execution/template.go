package execution

import (
	"fmt"
	"os"
	"path/filepath"

	"polyglot-rosetta/internal/domain/language"
)

// ResolveTemplatePath checks for a concept-specific template override.
// If not found, it falls back to the global default template directory.
func ResolveTemplatePath(conceptDir string, lang language.Language, globalDir string) (string, error) {
	ext := getExtension(lang)

	// 1. Check for local concept-specific override (e.g., conceptDir/templates/solution.go)
	localOverride := filepath.Join(conceptDir, "templates", fmt.Sprintf("solution.%s", ext))
	if _, err := os.Stat(localOverride); err == nil {
		return localOverride, nil
	}

	// 2. Fall back to global default template
	globalDefault := filepath.Join(globalDir, string(lang), fmt.Sprintf("default.%s", ext))
	if _, err := os.Stat(globalDefault); err == nil {
		return globalDefault, nil
	}

	return "", fmt.Errorf("no template found for language %s (checked local and global paths)", lang)
}

func getExtension(lang language.Language) string {
	switch lang {
	case language.Go:
		return "go"
	case language.Python:
		return "py"
	case language.Rust:
		return "rs"
	case language.TypeScript:
		return "ts"
	case language.Zig:
		return "zig"
	default:
		return "txt"
	}
}
