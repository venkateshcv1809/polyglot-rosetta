package run

import (
	"polyglot-rosetta/internal/api"
	"polyglot-rosetta/internal/domain/evaluation"
)

// Input source values describe how a native command obtained submitted code.
// Browser requests supply source directly and therefore use InputSourceInline.
const (
	InputSourceInline  = "inline"
	InputSourceStdin   = "stdin"
	InputSourceCatalog = "catalog"
)

// Result is the structured outcome of one run request.
type Result struct {
	SchemaVersion   int               `json:"schema_version"`
	ID              string            `json:"id"`
	Concept         string            `json:"concept,omitempty"`
	Language        string            `json:"language"`
	Mode            evaluation.Mode   `json:"mode"`
	InputSource     string            `json:"input_source,omitempty"`
	Status          evaluation.Status `json:"status"`
	TotalPassed     int               `json:"total_passed"`
	TotalFailed     int               `json:"total_failed"`
	DurationMs      int64             `json:"duration_ms"`
	BuildDurationMs *int64            `json:"build_duration_ms,omitempty"`
	TestResults     []TestResult      `json:"test_results"`
	Error           string            `json:"error,omitempty"`
}

// NewResult starts a run result on the current JSON contract.
func NewResult() Result {
	return Result{SchemaVersion: api.SchemaVersion, Status: evaluation.StatusError, TestResults: []TestResult{}}
}

// TestResult describes one selected test case without exposing its input.
type TestResult struct {
	ID             string                `json:"id"`
	Status         evaluation.TestStatus `json:"status"`
	Passed         bool                  `json:"passed"`
	Hidden         bool                  `json:"hidden"`
	Message        string                `json:"message,omitempty"`
	SolutionTimeNs int64                 `json:"solution_time_ns,omitempty"`
}
