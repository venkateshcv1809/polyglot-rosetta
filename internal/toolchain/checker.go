package toolchain

import (
	"polyglot-rosetta/internal/domain/language"
)

type ToolchainInfo struct {
	Language  language.Language `json:"language"`
	Available bool              `json:"available"`
	Version   string            `json:"version,omitempty"`
	Path      string            `json:"path,omitempty"`
}

type Checker interface {
	Check() (ToolchainInfo, error)
}
