package main

import (
	"os"
	"testing"

	generatecmd "polyglot-rosetta/internal/commands/generate"
)

func TestMainHandlesHelp(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	os.Args = []string{"generate", "--help"}

	main()
}

func TestGenerateEntrypoint(t *testing.T) {
	command := generatecmd.Command()
	if command.Name != "generate" {
		t.Fatalf("command name = %q, want %q", command.Name, "generate")
	}
	if err := generatecmd.Execute([]string{"--help"}); err != nil {
		t.Fatalf("help returned an error: %v", err)
	}
}
