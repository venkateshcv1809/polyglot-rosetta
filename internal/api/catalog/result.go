// Package catalog defines the structured result of making catalog content available.
package catalog

import (
	"polyglot-rosetta/internal/api"
	"polyglot-rosetta/internal/catalog"
)

// Status identifies whether catalog content is available.
type Status string

const (
	// StatusOK means the catalog snapshot is ready for evaluation.
	StatusOK Status = "ok"
	// StatusError means the request or synchronization failed.
	StatusError Status = "error"
)

// Result identifies the native problems tree or the browser's in-memory catalog.
// Only the field appropriate to the host is populated on success.
type Result struct {
	SchemaVersion int              `json:"schema_version"`
	Status        Status           `json:"status"`
	ConceptID     string           `json:"concept_id,omitempty"`
	Root          string           `json:"root,omitempty"`
	Catalog       *catalog.Catalog `json:"catalog,omitempty"`
	Error         string           `json:"error,omitempty"`
}

// NewResult starts a catalog result on the current JSON contract.
func NewResult(status Status) Result {
	return Result{SchemaVersion: api.SchemaVersion, Status: status}
}
