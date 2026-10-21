package run

import (
	"os/exec"

	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/toolchain"
)

// RequireToolchain checks the cached native support for a language, probing
// only when the complete availability snapshot is missing or invalid.
func RequireToolchain(lang language.Language) error {
	cachePath, err := toolchain.DefaultCachePath()
	if err != nil {
		return err
	}
	return toolchain.RequireAvailable(cachePath, lang, exec.LookPath, toolchain.InspectBinary)
}
