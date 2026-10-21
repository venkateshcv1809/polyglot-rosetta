package execution

import (
	"polyglot-rosetta/internal/domain/language"
)

// Status describes the overall outcome of an execution request.
type Status string

const (
	// StatusPassed indicates that every selected testcase passed.
	StatusPassed Status = "passed"
	// StatusFailed indicates that at least one testcase failed.
	StatusFailed Status = "failed"
	// StatusError indicates that evaluation could not complete.
	StatusError Status = "error"
	// StatusTimeout indicates that evaluation exceeded its deadline.
	StatusTimeout Status = "timeout"
)

// Request contains the native execution inputs for one solution.
type Request struct {
	Language   language.Language `json:"language"`
	SourceCode string            `json:"source_code"`
	IsSubmit   bool              `json:"is_submit"`
}
