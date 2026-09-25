package execution

import (
	"polyglot-rosetta/internal/domain/testcase"
)

// TestOutcome records the result of an individual test vector execution.
type TestOutcome struct {
	TestCaseID string
	Passed     bool
	Message    string
	Hidden     bool
}

// EvaluationReport summarizes the final scoring metrics for a submission run.
type EvaluationReport struct {
	Concept      string
	TotalTests   int
	PassedTests  int
	ScorePercent float64
	Outcomes     []TestOutcome
	PassedAll    bool
}

// AggregateResults compiles test outcomes into a unified evaluation score report.
func AggregateResults(suite *testcase.TestSuite, outcomes []TestOutcome) EvaluationReport {
	total := len(outcomes)
	passed := 0

	for _, outcome := range outcomes {
		if outcome.Passed {
			passed++
		}
	}

	var score float64
	if total > 0 {
		score = (float64(passed) / float64(total)) * 100.0
	}

	return EvaluationReport{
		Concept:      suite.Concept,
		TotalTests:   total,
		PassedTests:  passed,
		ScorePercent: score,
		Outcomes:     outcomes,
		PassedAll:    passed == total,
	}
}
