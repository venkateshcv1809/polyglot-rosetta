package execution

import (
	"testing"

	"polyglot-rosetta/internal/domain/testcase"
)

func TestAggregateResults(t *testing.T) {
	suite := &testcase.TestSuite{
		Concept: "two-sum",
		TestCases: []testcase.TestCase{
			{ID: "tc_1", Hidden: false},
			{ID: "tc_2", Hidden: true},
		},
	}

	outcomes := []TestOutcome{
		{TestCaseID: "tc_1", Passed: true, Hidden: false},
		{TestCaseID: "tc_2", Passed: false, Hidden: true, Message: "assertion failed"},
	}

	report := AggregateResults(suite, outcomes)

	if report.TotalTests != 2 {
		t.Errorf("expected 2 total tests, got %d", report.TotalTests)
	}
	if report.PassedTests != 1 {
		t.Errorf("expected 1 passed test, got %d", report.PassedTests)
	}
	if report.ScorePercent != 50.0 {
		t.Errorf("expected score 50.0, got %f", report.ScorePercent)
	}
	if report.PassedAll {
		t.Errorf("expected PassedAll to be false")
	}
}
