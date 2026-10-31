package execution

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	defaultCommandOutputLimit = 1 << 20
	defaultProcessWaitDelay   = 250 * time.Millisecond
)

// Command describes a child process executed by the evaluation engine.
type Command struct {
	Name           string
	Args           []string
	Dir            string
	Env            []string
	MaxOutputBytes int
}

// RunResult captures the outcome of a child process execution.
type RunResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	TimedOut bool
	// StdoutTruncated and StderrTruncated report that only the tail of the
	// corresponding stream was retained.
	StdoutTruncated bool
	StderrTruncated bool
	Error           error
}

// RunCommand executes a command with bounded output capture and a strict
// timeout. Environment entries supplement the current process environment.
func RunCommand(ctx context.Context, timeout time.Duration, command Command) RunResult {
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, command.Name, command.Args...)
	configureProcessGroup(cmd)
	cmd.Dir = command.Dir
	if len(command.Env) > 0 {
		cmd.Env = append(os.Environ(), command.Env...)
	}

	outputLimit := command.MaxOutputBytes
	if outputLimit <= 0 {
		outputLimit = defaultCommandOutputLimit
	}
	stdout := newTailBuffer(outputLimit)
	stderr := newTailBuffer(outputLimit)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := RunResult{
		Stdout:          stdout.String(),
		Stderr:          stderr.String(),
		StdoutTruncated: stdout.Truncated(),
		StderrTruncated: stderr.Truncated(),
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

type tailBuffer struct {
	data      []byte
	limit     int
	truncated bool
}

func newTailBuffer(limit int) tailBuffer {
	return tailBuffer{data: make([]byte, 0, limit), limit: limit}
}

func (b *tailBuffer) Write(data []byte) (int, error) {
	written := len(data)
	if len(data) >= b.limit {
		b.data = append(b.data[:0], data[len(data)-b.limit:]...)
		b.truncated = true
		return written, nil
	}

	overflow := len(b.data) + len(data) - b.limit
	if overflow > 0 {
		copy(b.data, b.data[overflow:])
		b.data = b.data[:len(b.data)-overflow]
		b.truncated = true
	}
	b.data = append(b.data, data...)
	return written, nil
}

func (b *tailBuffer) String() string {
	return string(b.data)
}

func (b *tailBuffer) Truncated() bool {
	return b.truncated
}
