package ui

import (
	"strings"
	"testing"

	"polyglot-rosetta/internal/domain/execution"
)

func TestFormatOutcome(t *testing.T) {
	passOutcome := execution.TestOutcome{
		TestCaseID: "tc_1",
		Passed:     true,
		Hidden:     false,
	}

	formatted := FormatOutcome(passOutcome)
	if !strings.Contains(formatted, "[ PASS ]") {
		t.Errorf("expected pass badge, got %s", formatted)
	}

	failOutcome := execution.TestOutcome{
		TestCaseID: "tc_2",
		Passed:     false,
		Hidden:     true,
	}

	formattedFail := FormatOutcome(failOutcome)
	if !strings.Contains(formattedFail, "[ FAIL ]") || !strings.Contains(formattedFail, "Hidden") {
		t.Errorf("expected fail badge and hidden label, got %s", formattedFail)
	}
}

func TestFormatReport(t *testing.T) {
	report := execution.EvaluationReport{
		Concept:      "two-sum",
		TotalTests:   2,
		PassedTests:  2,
		ScorePercent: 100.0,
		PassedAll:    true,
	}

	formatted := FormatReport(report)
	if !strings.Contains(formatted, "two-sum") || !strings.Contains(formatted, "100.0%") {
		t.Errorf("expected report to contain concept and score, got %s", formatted)
	}
}
