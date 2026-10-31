package evaluation

import (
	"polyglot-rosetta/internal/domain/language"
)

// Mode controls which testcases participate in an evaluation.
type Mode string

const (
	// ModePublic evaluates only cases intended for normal practice.
	ModePublic Mode = "public"
	// ModeAll evaluates every locally available testcase.
	ModeAll Mode = "all"
)

// IncludeHidden reports whether this mode selects hidden cases.
func (m Mode) IncludeHidden() bool {
	return m == ModeAll
}

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

// Request contains the source and language inputs for one evaluation.
type Request struct {
	Language   language.Language `json:"language"`
	SourceCode string            `json:"source_code"`
	Mode       Mode              `json:"mode"`
}
