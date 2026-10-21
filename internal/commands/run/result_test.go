package runcmd

import (
	"testing"
	"time"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/domain/execution"
	engine "polyglot-rosetta/internal/execution"
)

func TestInitialResult(t *testing.T) {
	result := initialResult(options{id: "c-12345678", language: "go", all: true})
	if result.ID != "c-12345678" || result.Language != "go" || result.Mode != runapi.ModeAll || result.Status != execution.StatusError || result.TestResults == nil {
		t.Fatalf("initial result = %+v", result)
	}
}

func TestResultFromEvaluation(t *testing.T) {
	for _, test := range []struct {
		name       string
		passedAll  bool
		wantStatus execution.Status
	}{
		{name: "passed", passedAll: true, wantStatus: execution.StatusPassed},
		{name: "failed", wantStatus: execution.StatusFailed},
	} {
		t.Run(test.name, func(t *testing.T) {
			passed := 1
			buildDuration := 1200 * time.Microsecond
			hiddenStatus := execution.TestWrongAnswer
			if test.passedAll {
				passed = 2
				hiddenStatus = execution.TestPassed
			}
			evaluation := engine.Evaluation{
				Duration: 1500 * time.Microsecond, BuildDuration: &buildDuration,
				Report: execution.EvaluationReport{
					TotalTests: 2, PassedTests: passed, PassedAll: test.passedAll,
					Outcomes: []execution.TestOutcome{
						{TestCaseID: "public", Status: execution.TestPassed, Passed: true, SolutionTimeNs: 123},
						{TestCaseID: "hidden", Status: hiddenStatus, Passed: test.passedAll, Hidden: true, SolutionTimeNs: 456},
					},
				},
			}
			result := resultFromEvaluation(initialResult(options{}), evaluation)
			if result.Status != test.wantStatus || result.DurationMs != 2 || result.BuildDurationMs == nil || *result.BuildDurationMs != 1 || result.TotalPassed != passed || result.TotalFailed != 2-passed {
				t.Fatalf("summary = %+v", result)
			}
			if len(result.TestResults) != 2 || result.TestResults[0].Status != execution.TestPassed || result.TestResults[0].SolutionTimeNs != 123 ||
				!result.TestResults[1].Hidden || result.TestResults[1].Status != hiddenStatus ||
				result.TestResults[1].Passed != test.passedAll || result.TestResults[1].SolutionTimeNs != 456 {
				t.Fatalf("test results = %+v", result.TestResults)
			}
		})
	}
}
