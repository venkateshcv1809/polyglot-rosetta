// Package probe defines the structured outcome of a local toolchain probe.
package probe

import "polyglot-rosetta/internal/api"

type Status string

const (
	// StatusOK means every required tool is available.
	StatusOK Status = "ok"
	// StatusUnavailable means at least one required tool is missing or unusable.
	StatusUnavailable Status = "unavailable"
	// StatusCleared means the saved toolchain snapshot was removed.
	StatusCleared Status = "cleared"
	// StatusError means the request could not be parsed.
	StatusError Status = "error"
)

// Result records the availability of the requested language toolchains.
type Result struct {
	SchemaVersion int         `json:"schema_version"`
	Status        Status      `json:"status"`
	Toolchains    []Toolchain `json:"toolchains"`
	Error         string      `json:"error,omitempty"`
}

// NewResult starts a probe result on the current JSON contract.
func NewResult(status Status) Result {
	return Result{SchemaVersion: api.SchemaVersion, Status: status, Toolchains: []Toolchain{}}
}

// Toolchain records the tools required by one language.
type Toolchain struct {
	Language  string `json:"language"`
	Available bool   `json:"available"`
	Tools     []Tool `json:"tools"`
}

// Tool records the result of locating and inspecting one executable.
type Tool struct {
	Binary    string `json:"binary"`
	Available bool   `json:"available"`
	Path      string `json:"path,omitempty"`
	Version   string `json:"version,omitempty"`
	Error     string `json:"error,omitempty"`
}
