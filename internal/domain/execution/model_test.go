package execution

import (
	"encoding/json"
	"testing"

	"polyglot-rosetta/internal/domain/language"
)

func TestRequestJSONSerialization(t *testing.T) {
	req := Request{
		Language:   language.Go,
		ConceptDir: "algorithms/two-sum",
		SourceCode: "package main\nfunc Add(a, b int) int { return a + b }",
		IsSubmit:   true,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal Request: %v", err)
	}

	var decoded Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Request: %v", err)
	}

	if decoded.Language != req.Language || decoded.ConceptDir != req.ConceptDir || decoded.SourceCode != req.SourceCode || decoded.IsSubmit != req.IsSubmit {
		t.Errorf("Decoded request does not match original: %+v", decoded)
	}
}

func TestResultInstantiationAndStatus(t *testing.T) {
	res := Result{
		Language:    language.Python,
		Status:      StatusPassed,
		ExitCode:    0,
		Timeout:     false,
		TotalPassed: 5,
		TotalFailed: 0,
		DurationMs:  125,
		TestResults: []TestCaseResult{
			{
				ID:         "test_1",
				Passed:     true,
				Expected:   "4",
				Actual:     "4",
				DurationMs: 20,
			},
		},
		Stdout: "All tests passed successfully.",
	}

	if res.Status != StatusPassed {
		t.Errorf("Expected status to be 'passed', got %s", res.Status)
	}

	if res.TotalPassed != 5 || res.TotalFailed != 0 {
		t.Errorf("Incorrect aggregate counts: passed=%d, failed=%d", res.TotalPassed, res.TotalFailed)
	}

	if len(res.TestResults) != 1 || !res.TestResults[0].Passed {
		t.Errorf("Test results array not populated or recorded incorrectly")
	}
}
