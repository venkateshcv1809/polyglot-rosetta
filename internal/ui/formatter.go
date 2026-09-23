package ui

import (
	"fmt"

	"polyglot-rosetta/internal/domain/execution"
)

// FormatOutcome returns a styled status string with color-coded badges for CLI output.
func FormatOutcome(outcome execution.TestOutcome) string {
	badge := "[ PASS ]"
	colorCode := "\033[32m" // Green for success
	reset := "\033[0m"

	if !outcome.Passed {
		badge = "[ FAIL ]"
		colorCode = "\033[31m" // Red for failure
	}

	visibility := "Public"
	if outcome.Hidden {
		visibility = "Hidden"
	}

	return fmt.Sprintf("%s%s%s (%s) -> %s", colorCode, badge, reset, visibility, outcome.TestCaseID)
}

// FormatReport generates a summary block for the final evaluation report.
func FormatReport(report execution.EvaluationReport) string {
	status := "\033[32mSUCCESS\033[0m"
	if !report.PassedAll {
		status = "\033[31mFAILED\033[0m"
	}

	return fmt.Sprintf("\nConcept: %s\nStatus: %s\nScore: %.1f%% (%d/%d passed)\n",
		report.Concept, status, report.ScorePercent, report.PassedTests, report.TotalTests)
}
