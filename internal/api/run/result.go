package run

import "polyglot-rosetta/internal/domain/execution"

// Result is the structured outcome of one run request.
type Result struct {
	ID              string           `json:"id"`
	Concept         string           `json:"concept,omitempty"`
	Language        string           `json:"language"`
	Mode            Mode             `json:"mode"`
	InputSource     string           `json:"input_source,omitempty"`
	Status          execution.Status `json:"status"`
	TotalPassed     int              `json:"total_passed"`
	TotalFailed     int              `json:"total_failed"`
	DurationMs      int64            `json:"duration_ms"`
	BuildDurationMs *int64           `json:"build_duration_ms,omitempty"`
	TestResults     []TestResult     `json:"test_results"`
	Error           string           `json:"error,omitempty"`
}

// TestResult describes one selected test case without exposing its input.
type TestResult struct {
	ID             string               `json:"id"`
	Status         execution.TestStatus `json:"status"`
	Passed         bool                 `json:"passed"`
	Hidden         bool                 `json:"hidden"`
	Message        string               `json:"message,omitempty"`
	SolutionTimeNs int64                `json:"solution_time_ns,omitempty"`
}
