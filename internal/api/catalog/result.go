// Package catalog defines the structured result of native catalog synchronization.
package catalog

// Status identifies whether a catalog snapshot was activated.
type Status string

const (
	// StatusOK means the catalog snapshot is ready for native evaluation.
	StatusOK Status = "ok"
	// StatusError means the request or synchronization failed.
	StatusError Status = "error"
)

// Result identifies the activated problems tree or the reason it is unavailable.
type Result struct {
	Status Status `json:"status"`
	Root   string `json:"root,omitempty"`
	Error  string `json:"error,omitempty"`
}
