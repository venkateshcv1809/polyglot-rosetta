package schema

import "fmt"

// Language represents a supported programming language.
type Language string

const (
	LanguageGo         Language = "go"
	LanguagePython     Language = "python"
	LanguageTypeScript Language = "typescript"
	LanguageRust       Language = "rust"
	LanguageZig        Language = "zig"
)

// SupportedLanguages returns a slice of all supported language enums.
var SupportedLanguages = []Language{
	LanguageGo,
	LanguagePython,
	LanguageTypeScript,
	LanguageRust,
	LanguageZig,
}

// Validate checks if a language string is supported.
func (l Language) Validate() error {
	switch l {
	case LanguageGo, LanguagePython, LanguageTypeScript, LanguageRust, LanguageZig:
		return nil
	default:
		return fmt.Errorf("unsupported language %q", l)
	}
}

// FileExtension returns the main file extension for the language.
func (l Language) FileExtension() string {
	switch l {
	case LanguageGo:
		return "go"
	case LanguagePython:
		return "py"
	case LanguageTypeScript:
		return "ts"
	case LanguageRust:
		return "rs"
	case LanguageZig:
		return "zig"
	default:
		return ""
	}
}
