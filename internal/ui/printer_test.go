package ui

import (
	"strings"
	"testing"
	"time"

	"polyglot-rosetta/internal/domain/execution"
)

func TestFormatDetailedSummary(t *testing.T) {
	report := execution.EvaluationReport{
		Concept:      "two-sum",
		TotalTests:   4,
		PassedTests:  3,
		ScorePercent: 75.0,
		PassedAll:    false,
	}

	duration := 150 * time.Millisecond
	summary := FormatDetailedSummary(report, duration)

	if !strings.Contains(summary, "two-sum") {
		t.Errorf("expected summary to contain concept name, got %s", summary)
	}
	if !strings.Contains(summary, "75.0%") {
		t.Errorf("expected summary to contain score percentage, got %s", summary)
	}
	if !strings.Contains(summary, "150ms") {
		t.Errorf("expected summary to contain elapsed time, got %s", summary)
	}
}
