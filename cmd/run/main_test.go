package main

import (
	"errors"
	"flag"
	"os"
	"testing"

	runcmd "polyglot-rosetta/internal/commands/run"
)

func TestMainHandlesHelp(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	os.Args = []string{"run", "--help"}

	main()
}

func TestRunEntrypoint(t *testing.T) {
	command := runcmd.Command()
	if command.Name != "run" {
		t.Fatalf("command name = %q, want %q", command.Name, "run")
	}
	if err := runcmd.Execute([]string{"--help"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("help error = %v, want flag.ErrHelp", err)
	}
}
