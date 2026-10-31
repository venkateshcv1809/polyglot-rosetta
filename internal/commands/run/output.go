package runcmd

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/domain/evaluation"
)

//go:embed output/*.tmpl
var outputFiles embed.FS

type testDisplay struct {
	Status   string
	Color    string
	Case     string
	Solution string
	Reason   string
}

type resultDisplay struct {
	Concept       string
	Language      string
	Cases         evaluation.Mode
	Status        string
	Passed        int
	Total         int
	Score         float64
	Duration      time.Duration
	BuildDuration *time.Duration
	Border        string
	StatusHeading string
	CaseHeading   string
	TimeHeading   string
	ReasonHeading string
	HasReasons    bool
	Tests         []testDisplay
}

func writeResult(output io.Writer, format string, result runapi.Result) error {
	if format == "json" {
		if err := json.NewEncoder(output).Encode(result); err != nil {
			return fmt.Errorf("write JSON result: %w", err)
		}
		return nil
	}
	return writeEvaluationEnd(output, result)
}

func writeEvaluationEnd(output io.Writer, result runapi.Result) error {
	if result.Error != "" {
		return nil
	}
	view := resultDisplay{
		Concept: result.Concept, Language: result.Language, Cases: result.Mode, Passed: result.TotalPassed,
		Total:    result.TotalPassed + result.TotalFailed,
		Duration: time.Duration(result.DurationMs) * time.Millisecond,
		Tests:    make([]testDisplay, 0, len(result.TestResults)),
	}
	if result.BuildDurationMs != nil {
		buildDuration := time.Duration(*result.BuildDurationMs) * time.Millisecond
		view.BuildDuration = &buildDuration
	}
	view.Status = "\033[32m[ PASSED ]\033[0m"
	if result.Status != evaluation.StatusPassed {
		view.Status = "\033[31m[ FAILED ]\033[0m"
	}
	if view.Total > 0 {
		view.Score = float64(view.Passed) / float64(view.Total) * 100
	}
	for _, test := range result.TestResults {
		entry := testDisplay{
			Status: "[ PASS ]", Color: "\033[32m", Case: "Public " + test.ID,
			Solution: time.Duration(test.SolutionTimeNs).String(), Reason: test.Message,
		}
		switch test.Status {
		case evaluation.TestCrashed:
			entry.Status, entry.Color, entry.Solution = "[ CRASH ]", "\033[31m", "-"
		case evaluation.TestTimeout:
			entry.Status, entry.Color, entry.Solution = "[ TIMEOUT ]", "\033[33m", "-"
		case evaluation.TestWrongAnswer:
			entry.Status, entry.Color = "[ FAIL ]", "\033[31m"
		default:
			if !test.Passed {
				entry.Status, entry.Color = "[ FAIL ]", "\033[31m"
			}
		}
		if test.Hidden {
			entry.Case = "Hidden " + test.ID
		}
		if entry.Reason != "" {
			view.HasReasons = true
		}
		view.Tests = append(view.Tests, entry)
	}
	formatResultsTable(&view)
	return renderText(output, "result.tmpl", view)
}

func formatResultsTable(view *resultDisplay) {
	widths := []int{len("Status"), len("Case"), len("Solution")}
	if view.HasReasons {
		widths = append(widths, len("Reason"))
	}
	for _, test := range view.Tests {
		for column, value := range []string{test.Status, test.Case, test.Solution} {
			widths[column] = max(widths[column], utf8.RuneCountInString(value))
		}
		if view.HasReasons {
			widths[3] = max(widths[3], utf8.RuneCountInString(test.Reason))
		}
	}
	var border strings.Builder
	for _, width := range widths {
		border.WriteByte('+')
		border.WriteString(strings.Repeat("-", width+2))
	}
	border.WriteByte('+')
	view.Border = border.String()
	view.StatusHeading = paddedCell("Status", widths[0])
	view.CaseHeading = paddedCell("Case", widths[1])
	view.TimeHeading = paddedCell("Solution", widths[2])
	if view.HasReasons {
		view.ReasonHeading = paddedCell("Reason", widths[3])
	}
	for i := range view.Tests {
		entry := &view.Tests[i]
		entry.Status = entry.Color + paddedCell(entry.Status, widths[0]) + "\033[0m"
		entry.Case = paddedCell(entry.Case, widths[1])
		entry.Solution = paddedCell(entry.Solution, widths[2])
		if view.HasReasons {
			entry.Reason = paddedCell(entry.Reason, widths[3])
		}
	}
}

func paddedCell(value string, width int) string {
	return value + strings.Repeat(" ", width-utf8.RuneCountInString(value))
}

func renderText(output io.Writer, name string, data any) error {
	tmpl, err := template.New("run-output").Option("missingkey=error").ParseFS(outputFiles, "output/*.tmpl")
	if err != nil {
		return fmt.Errorf("parse run output templates: %w", err)
	}
	if err := tmpl.ExecuteTemplate(output, name, data); err != nil {
		return fmt.Errorf("render run %s: %w", name, err)
	}
	return nil
}
