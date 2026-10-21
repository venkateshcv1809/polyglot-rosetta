package cli

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func TestCommandFailure(t *testing.T) {
	cause := errors.New("target directory already exists")
	err := CommandFailure("Concept not created", cause)
	if !errors.Is(err, cause) {
		t.Fatalf("wrapped error = %v, want cause", err)
	}
	if CommandFailure("Concept not created", nil) != nil {
		t.Fatal("nil cause must remain nil")
	}

	var output bytes.Buffer
	if code := ExitCode(fmt.Errorf("generate concept: %w", err), &output); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	want := "\nERROR:\n  Concept not created\n  Reason:  target directory already exists\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

func TestReportedErrorKeepsExitStatusWithoutDuplicateOutput(t *testing.T) {
	cause := errors.New("run failed")
	err := ReportedError(cause)
	if !errors.Is(err, cause) {
		t.Fatalf("reported error = %v, want cause", err)
	}
	if ReportedError(nil) != nil {
		t.Fatal("nil cause must remain nil")
	}

	var stderr bytes.Buffer
	if code := ExitCode(fmt.Errorf("run: %w", err), &stderr); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr = %q", stderr.String())
	}
}
