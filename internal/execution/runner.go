package execution

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// RunResult captures the outcome of an isolated process execution.
type RunResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	TimedOut bool
	Error    error
}

// RunWithTimeout executes a command with isolated stdout/stderr buffer capture and strict timeout limits.
func RunWithTimeout(ctx context.Context, timeout time.Duration, name string, args ...string) RunResult {
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, name, args...)

	// Independent memory buffers to prevent cross-contamination and channel deadlocks
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := RunResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if execCtx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.Error = fmt.Errorf("execution exceeded hard limit of %v", timeout)
		result.ExitCode = -1
		return result
	}

	if err != nil {
		result.Error = err
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	return result
}
