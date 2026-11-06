package main

import (
	"os"
	"strings"
	"testing"

	sitecmd "polyglot-rosetta/internal/commands/site"
)

func TestMainHandlesHelp(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	os.Args = []string{"site", "--help"}

	main()
}

func TestSiteEntrypoint(t *testing.T) {
	command := sitecmd.Command()
	if command.Name != "site" {
		t.Fatalf("command name = %q, want %q", command.Name, "site")
	}
	if err := sitecmd.Execute([]string{"--help"}); err != nil {
		t.Fatalf("help returned an error: %v", err)
	}
}

func TestExecute_MissingCommand(t *testing.T) {
	if err := sitecmd.Execute(nil); err == nil || !strings.Contains(err.Error(), "missing command") {
		t.Fatalf("error = %v, want missing command", err)
	}
}
