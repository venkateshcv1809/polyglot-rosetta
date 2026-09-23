package execution

import (
	"context"
	"testing"
	"time"
)

func TestRunWithTimeout_Success(t *testing.T) {
	// Use 'echo' or a quick command depending on platform availability,
	// or standard go command for safety.
	result := RunWithTimeout(context.Background(), 2*time.Second, "go", "version")

	if result.TimedOut {
		t.Errorf("expected command not to time out")
	}
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestRunWithTimeout_TimeoutExceeded(t *testing.T) {
	// Use sleep to simulate a long-running command with a very tight timeout
	// On Unix systems, sleep is universally available.
	result := RunWithTimeout(context.Background(), 50*time.Millisecond, "sleep", "2")

	if !result.TimedOut {
		t.Errorf("expected TimedOut to be true")
	}
	if result.Error == nil {
		t.Errorf("expected timeout error, got nil")
	}
}

func TestRunWithTimeout_StreamSeparation(t *testing.T) {
	// Execute a command that writes to both stdout (or prints version) and check separation.
	result := RunWithTimeout(context.Background(), 2*time.Second, "go", "version")

	if result.Stdout == "" {
		t.Errorf("expected stdout buffer to capture output, got empty string")
	}
	if result.Stderr != "" {
		t.Errorf("expected stderr buffer to be empty for a successful go version call, got %s", result.Stderr)
	}
}
