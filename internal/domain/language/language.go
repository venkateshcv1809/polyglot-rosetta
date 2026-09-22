package language

import (
	"errors"
	"strings"
)

type Language string

const (
	Go         Language = "go"
	Python     Language = "python"
	Rust       Language = "rust"
	TypeScript Language = "typescript"
	Zig        Language = "zig"
)

var ErrUnsupportedLanguage = errors.New("unsupported language target")

func AllSupported() []Language {
	return []Language{Go, Python, Rust, TypeScript, Zig}
}

func Parse(raw string) (Language, error) {
	normalized := Language(strings.ToLower(strings.TrimSpace(raw)))

	switch normalized {
	case Go, Python, Rust, TypeScript, Zig:
		return normalized, nil
	default:
		return "", ErrUnsupportedLanguage
	}
}

func IsValid(raw string) bool {
	_, err := Parse(raw)
	return err == nil
}
