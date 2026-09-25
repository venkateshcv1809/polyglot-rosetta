package execution

import (
	"polyglot-rosetta/internal/domain/language"
)

type Status string

const (
	StatusPassed  Status = "passed"
	StatusFailed  Status = "failed"
	StatusError   Status = "error"
	StatusTimeout Status = "timeout"
)

type Request struct {
	Language   language.Language `json:"language"`
	ConceptDir string            `json:"concept_dir"`
	SourceCode string            `json:"source_code"`
	IsSubmit   bool              `json:"is_submit"`
}

type Result struct {
	Language    language.Language `json:"language"`
	Status      Status            `json:"status"`
	ExitCode    int               `json:"exit_code"`
	Timeout     bool              `json:"timeout"`
	TotalPassed int               `json:"total_passed"`
	TotalFailed int               `json:"total_failed"`
	DurationMs  int64             `json:"duration_ms"`
	TestResults []TestCaseResult  `json:"test_results"`
	Stdout      string            `json:"stdout,omitempty"`
	Stderr      string            `json:"stderr,omitempty"`
}

type TestCaseResult struct {
	ID         string `json:"id"`
	Passed     bool   `json:"passed"`
	Expected   string `json:"expected"`
	Actual     string `json:"actual"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}
