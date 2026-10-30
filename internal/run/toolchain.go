package run

import (
	"os/exec"

	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/toolchain"
)

// EnsureAvailable checks the cached native support for a language, probing
// only when the complete availability snapshot is missing or invalid.
func EnsureAvailable(lang language.Language) error {
	cachePath, err := toolchain.DefaultCachePath()
	if err != nil {
		return err
	}
	return toolchain.EnsureAvailable(cachePath, lang, exec.LookPath, toolchain.InspectBinary)
}
