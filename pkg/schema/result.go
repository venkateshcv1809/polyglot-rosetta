package schema

import "time"

type TestCaseResult struct {
	ID        int           `json:"id"`
	Passed    bool          `json:"passed"`
	Duration  time.Duration `json:"duration_ns"`
	Actual    string        `json:"actual,omitempty"`
	Expected  string        `json:"expected,omitempty"`
	Error     string        `json:"error,omitempty"`
}

type LanguageResult struct {
	Language  Language         `json:"language"`
	Passed    bool             `json:"passed"`
	Duration  time.Duration    `json:"total_duration_ns"`
	TestCases []TestCaseResult `json:"test_cases"`
}

type ConceptResult struct {
	ConceptID string           `json:"concept_id"`
	EvaluatedAt time.Time      `json:"evaluated_at"`
	Languages []LanguageResult `json:"languages"`
}
