package execution

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunCommandSuccess(t *testing.T) {
	result := RunCommand(context.Background(), 2*time.Second, Command{Name: "go", Args: []string{"version"}})

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

func TestRunCommandTimeoutExceeded(t *testing.T) {
	command := helperCommand("sleep")
	result := RunCommand(context.Background(), 50*time.Millisecond, command)

	if !result.TimedOut {
		t.Errorf("expected TimedOut to be true")
	}
	if result.Error == nil {
		t.Errorf("expected timeout error, got nil")
	}
}

func TestRunCommandStreamSeparation(t *testing.T) {
	result := RunCommand(context.Background(), 2*time.Second, Command{Name: "go", Args: []string{"version"}})

	if result.Stdout == "" {
		t.Errorf("expected stdout buffer to capture output, got empty string")
	}
	if result.Stderr != "" {
		t.Errorf("expected stderr buffer to be empty for a successful go version call, got %s", result.Stderr)
	}
}

func TestRunCommandUsesDirectoryAndEnvironment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	source := `package main
import ("fmt"; "os")
func main() { fmt.Printf("%s:%s", os.Getenv("PROSETTA_TEST"), mustGetwd()) }
func mustGetwd() string { dir, _ := os.Getwd(); return dir }
`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	result := RunCommand(context.Background(), 5*time.Second, Command{
		Name: "go",
		Args: []string{"run", "main.go"},
		Dir:  dir,
		Env:  []string{"PROSETTA_TEST=ready"},
	})
	if result.Error != nil {
		t.Fatalf("RunCommand: %v\n%s", result.Error, result.Stderr)
	}
	if !strings.HasPrefix(result.Stdout, "ready:"+dir) {
		t.Fatalf("stdout = %q", result.Stdout)
	}
}

func TestRunCommandBoundsCapturedOutput(t *testing.T) {
	command := helperCommand("output")
	command.MaxOutputBytes = 4
	result := RunCommand(context.Background(), 2*time.Second, command)
	if result.Error != nil {
		t.Fatalf("RunCommand: %v", result.Error)
	}
	if result.Stdout != "6789" || !result.StdoutTruncated {
		t.Fatalf("stdout = %q, truncated = %t", result.Stdout, result.StdoutTruncated)
	}
	if result.Stderr != "ghij" || !result.StderrTruncated {
		t.Fatalf("stderr = %q, truncated = %t", result.Stderr, result.StderrTruncated)
	}
}

func helperCommand(mode string) Command {
	return Command{
		Name: os.Args[0],
		Args: []string{"-test.run=TestRunnerHelperProcess", "--", mode},
		Env:  []string{"GO_WANT_RUNNER_HELPER=1"},
	}
}

func TestRunnerHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "1" {
		return
	}
	mode := os.Args[len(os.Args)-1]
	switch mode {
	case "sleep":
		time.Sleep(5 * time.Second)
	case "output":
		fmt.Fprint(os.Stdout, "0123456789")
		fmt.Fprint(os.Stderr, "abcdefghij")
	case "spawn":
		command := exec.Command(os.Args[0], "-test.run=TestRunnerHelperProcess", "--", "sleep")
		command.Env = append(os.Environ(), "GO_WANT_RUNNER_HELPER=1")
		if err := command.Run(); err != nil {
			os.Exit(1)
		}
	default:
		os.Exit(2)
	}
	os.Exit(0)
}
