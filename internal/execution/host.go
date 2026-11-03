package execution

import (
	"context"
	"fmt"
	"time"

	"polyglot-rosetta/internal/api"
	executionapi "polyglot-rosetta/internal/api/execution"
)

// Host supplies only the operations that depend on an execution environment.
// The evaluator owns source generation, ordering, result parsing and grading.
type Host interface {
	BuildProgram(context.Context, executionapi.Program) (executionapi.BuildResult, error)
	RunCase(context.Context, executionapi.CaseRequest) (executionapi.CaseResult, error)
	DisposeProgram(context.Context, string) error
}

func validateBuild(program executionapi.Program, result executionapi.BuildResult) error {
	if result.SchemaVersion != api.SchemaVersion || result.SessionID != program.SessionID {
		return fmt.Errorf("invalid build result: schema or session identity does not match")
	}
	if result.BuildDurationMs != nil && (*result.BuildDurationMs < 0 || *result.BuildDurationMs > int64((1<<63-1)/time.Millisecond) || !program.BuildRequired) {
		return fmt.Errorf("invalid build result: unexpected or negative build duration")
	}
	if program.BuildRequired && result.Error == "" && result.BuildDurationMs == nil {
		return fmt.Errorf("invalid build result: compiled programs require build_duration_ms")
	}
	return nil
}

func processResult(request executionapi.CaseRequest, result executionapi.CaseResult) (RunResult, error) {
	if result.SchemaVersion != api.SchemaVersion || result.SessionID != request.SessionID || result.Index != request.Index {
		return RunResult{}, fmt.Errorf("invalid case result: schema, session or case identity does not match")
	}
	if result.ExitCode < -1 {
		return RunResult{}, fmt.Errorf("invalid case result: exit_code must be -1 or nonnegative")
	}
	stdout, stderr := newTailBuffer(request.MaxOutputBytes), newTailBuffer(request.MaxOutputBytes)
	_, _ = stdout.Write([]byte(result.Stdout))
	_, _ = stderr.Write([]byte(result.Stderr))
	process := RunResult{
		Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: result.ExitCode, TimedOut: result.TimedOut,
		StdoutTruncated: result.StdoutTruncated || stdout.Truncated(),
		StderrTruncated: result.StderrTruncated || stderr.Truncated(),
	}
	if result.Error != "" {
		process.Error = fmt.Errorf("%s", result.Error)
	} else if result.ExitCode != 0 {
		process.Error = fmt.Errorf("solution exited with code %d", result.ExitCode)
	}
	return process, nil
}

func durationMilliseconds(duration time.Duration) int64 {
	return duration.Round(time.Millisecond).Milliseconds()
}
