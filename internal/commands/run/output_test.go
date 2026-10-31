package runcmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/domain/evaluation"
)

func TestWriteEvaluationText(t *testing.T) {
	var output bytes.Buffer
	buildDurationMs := int64(125)
	result := runapi.Result{
		Concept: "fundamentals/hello-world", Language: "go", Mode: evaluation.ModePublic,
		InputSource: "stdin", Status: evaluation.StatusFailed, TotalFailed: 1, DurationMs: 150, BuildDurationMs: &buildDurationMs,
		TestResults: []runapi.TestResult{{ID: "case-1", Status: evaluation.TestWrongAnswer, Hidden: true, Message: "wrong result", SolutionTimeNs: 1200}},
	}
	if err := writeResult(&output, "text", result); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{
		"Results:\n", "| Status", "[ FAIL ]", "Hidden case-1",
		"Reason", "wrong result", "1.2µs", "Summary:\n", "Concept:      fundamentals/hello-world", "Language:     go", "Cases:        public", "0/1 tests passed", "Build Time:   125ms", "Elapsed Time: 150ms",
	} {
		if !strings.Contains(output.String(), text) {
			t.Fatalf("missing %q in %q", text, output.String())
		}
	}
	if strings.Contains(output.String(), "Run:") || strings.Contains(output.String(), "Source:") || strings.Contains(output.String(), "\n---\n") || strings.Count(output.String(), "\n  +") != 3 {
		t.Fatalf("unexpected run layout: %q", output.String())
	}
}

func TestWriteResultJSON(t *testing.T) {
	var output bytes.Buffer
	buildDurationMs := int64(125)
	result := runapi.Result{ID: "c-12345678", Status: evaluation.StatusError, Error: "build failed", BuildDurationMs: &buildDurationMs}
	if err := writeResult(&output, "json", result); err != nil {
		t.Fatal(err)
	}
	var decoded runapi.Result
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID != result.ID || decoded.Status != result.Status || decoded.Error != result.Error || decoded.BuildDurationMs == nil || *decoded.BuildDurationMs != buildDurationMs {
		t.Fatalf("decoded = %+v", decoded)
	}
}

func TestWriteEvaluationOmitsSummaryOnError(t *testing.T) {
	var output bytes.Buffer
	if err := writeResult(&output, "text", runapi.Result{
		Concept: "fundamentals/hello-world", Status: evaluation.StatusError, Error: "build failed",
	}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "Summary:") {
		t.Fatalf("unexpected summary: %q", output.String())
	}
}

func TestWriteEvaluationPassedText(t *testing.T) {
	var output bytes.Buffer
	result := runapi.Result{
		Concept: "two-sum", Language: "python", Mode: evaluation.ModeAll, InputSource: "catalog",
		Status: evaluation.StatusPassed, TotalPassed: 1, DurationMs: 42,
		TestResults: []runapi.TestResult{{ID: "case-1", Status: evaluation.TestPassed, Passed: true, SolutionTimeNs: 900}},
	}
	if err := writeResult(&output, "text", result); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Cases:        all", "[ PASS ]", "Public case-1", "| Solution", "900ns", "Language:     python", "100.0% (1/1 tests passed)", "Build Time:   not required", "Elapsed Time: 42ms"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("missing %q in %q", want, output.String())
		}
	}
	if strings.Contains(output.String(), "| Reason") {
		t.Fatalf("unexpected reason column for passing cases: %q", output.String())
	}
}

func TestWriteEvaluationShowsCrashAndTimeoutWithoutSolutionTime(t *testing.T) {
	var output bytes.Buffer
	result := runapi.Result{
		Concept: "two-sum", Status: evaluation.StatusFailed, TotalFailed: 2,
		TestResults: []runapi.TestResult{
			{ID: "crash", Status: evaluation.TestCrashed, Message: "solution exited with code 2"},
			{ID: "timeout", Status: evaluation.TestTimeout, Message: "solution exceeded the time limit"},
		},
	}
	if err := writeResult(&output, "text", result); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"[ CRASH ]", "[ TIMEOUT ]", "| Reason", "solution exited with code 2", "solution exceeded the time limit"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("missing %q in %q", want, output.String())
		}
	}
	if strings.Contains(output.String(), "| 0s") {
		t.Fatalf("unexpected solution time for unfinished calls: %q", output.String())
	}
}

func TestResultTableAlignsVariableWidthCells(t *testing.T) {
	var output bytes.Buffer
	result := runapi.Result{
		Concept: "two-sum", Status: evaluation.StatusFailed, TotalPassed: 1, TotalFailed: 1,
		TestResults: []runapi.TestResult{
			{ID: "1", Status: evaluation.TestPassed, Passed: true, SolutionTimeNs: 200},
			{ID: "long-case-identifier", Hidden: true, Status: evaluation.TestCrashed, Message: "solution exited with code 2"},
		},
	}
	if err := writeResult(&output, "text", result); err != nil {
		t.Fatal(err)
	}
	stripColor := strings.NewReplacer("\033[31m", "", "\033[32m", "", "\033[33m", "", "\033[0m", "")
	width := 0
	rows := 0
	for _, line := range strings.Split(stripColor.Replace(output.String()), "\n") {
		if !strings.HasPrefix(line, "  +") && !strings.HasPrefix(line, "  |") {
			continue
		}
		if width == 0 {
			width = len(line)
		}
		if len(line) != width {
			t.Fatalf("misaligned result table line %q (want width %d)", line, width)
		}
		rows++
	}
	if rows != 6 {
		t.Fatalf("table has %d lines, want 6: %q", rows, output.String())
	}
}
