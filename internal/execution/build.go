package execution

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const minimumBuildTimeout = 60 * time.Second

// buildProgram builds the solution and harness once for all selected cases.
// Duration covers the compiler process, including failed or timed-out attempts.
// Interpreted languages have no command and therefore no build duration.
func buildProgram(ctx context.Context, command *Command, timeout time.Duration) (*time.Duration, error) {
	if command == nil {
		return nil, nil
	}
	started := time.Now()
	result := RunCommand(ctx, max(timeout, minimumBuildTimeout), *command)
	duration := time.Since(started)
	if result.Error != nil {
		return &duration, buildError(*command, result)
	}
	return &duration, nil
}

func buildError(command Command, result RunResult) error {
	if result.TimedOut {
		return fmt.Errorf("build command %q timed out", command.Name)
	}
	details := strings.TrimSpace(result.Stderr)
	if details == "" {
		details = strings.TrimSpace(result.Stdout)
	}
	if details == "" {
		return fmt.Errorf("build command %q failed: %w", command.Name, result.Error)
	}
	if result.StdoutTruncated || result.StderrTruncated {
		details = "[output truncated; showing tail]\n" + details
	}
	return fmt.Errorf("build command %q failed: %w\n%s", command.Name, result.Error, details)
}
