package ui

import (
	"fmt"
	"time"

	"polyglot-rosetta/internal/domain/execution"
)

// FormatDetailedSummary generates a comprehensive execution summary block including duration and pass rates.
func FormatDetailedSummary(report execution.EvaluationReport, duration time.Duration) string {
	statusBadge := "\033[32m[ PASSED ]\033[0m"
	if !report.PassedAll {
		statusBadge = "\033[31m[ FAILED ]\033[0m"
	}

	return fmt.Sprintf(
		"\n────────────────────────────────────────\n"+
			"  Execution Summary: %s\n"+
			"────────────────────────────────────────\n"+
			"  • Concept:      %s\n"+
			"  • Status:       %s\n"+
			"  • Pass Rate:    %.1f%% (%d/%d tests passed)\n"+
			"  • Elapsed Time: %v\n"+
			"────────────────────────────────────────\n",
		report.Concept,
		report.Concept,
		statusBadge,
		report.ScorePercent,
		report.PassedTests,
		report.TotalTests,
		duration.Round(time.Millisecond),
	)
}
