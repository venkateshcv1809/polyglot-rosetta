package main

import (
	"errors"
	"flag"
	"os"
	"testing"

	probecmd "polyglot-rosetta/internal/commands/probe"
)

func TestMainHandlesHelp(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	os.Args = []string{"probe", "--help"}

	main()
}

func TestProbeEntrypoint(t *testing.T) {
	command := probecmd.Command()
	if command.Name != "probe" {
		t.Fatalf("command name = %q, want %q", command.Name, "probe")
	}
	if err := probecmd.Execute([]string{"--help"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("help error = %v, want flag.ErrHelp", err)
	}
}
