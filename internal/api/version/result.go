// Package version defines the structured result of the CLI version command.
package version

import "polyglot-rosetta/internal/api"

type Status string

const (
	StatusOK    Status = "ok"
	StatusError Status = "error"
)

type Result struct {
	SchemaVersion int    `json:"schema_version"`
	Status        Status `json:"status"`
	Version       string `json:"version,omitempty"`
	Error         string `json:"error,omitempty"`
}

// NewResult starts a version result on the current JSON contract.
func NewResult(status Status) Result {
	return Result{SchemaVersion: api.SchemaVersion, Status: status}
}
