// Package language defines the programming languages supported by Prosetta.
package language

import (
	"errors"
	"strings"
)

// Language identifies a supported solution language.
type Language string

const (
	Go         Language = "go"
	Python     Language = "python"
	Rust       Language = "rust"
	TypeScript Language = "typescript"
	Zig        Language = "zig"
)

// ErrUnsupportedLanguage is returned when a language name is not supported.
var ErrUnsupportedLanguage = errors.New("unsupported language target")

// AllSupported returns every supported language in display order.
func AllSupported() []Language {
	return []Language{Go, Python, Rust, TypeScript, Zig}
}

// Parse normalizes and validates a language name.
func Parse(raw string) (Language, error) {
	normalized := Language(strings.ToLower(strings.TrimSpace(raw)))

	switch normalized {
	case Go, Python, Rust, TypeScript, Zig:
		return normalized, nil
	default:
		return "", ErrUnsupportedLanguage
	}
}

// IsValid reports whether raw identifies a supported language.
func IsValid(raw string) bool {
	_, err := Parse(raw)
	return err == nil
}

// FileExtension returns the conventional source-file extension for the
// language, without a leading dot.
func (l Language) FileExtension() string {
	switch l {
	case Go:
		return "go"
	case Python:
		return "py"
	case Rust:
		return "rs"
	case TypeScript:
		return "ts"
	case Zig:
		return "zig"
	default:
		return ""
	}
}
