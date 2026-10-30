package listcmd

import (
	"errors"
	"flag"
	"testing"
)

func TestCommand(t *testing.T) {
	command := Command()
	if command.Name != commandName || command.Summary != commandSummary || command.Run == nil {
		t.Fatalf("list command = %#v", command)
	}
}

func TestParseOptions(t *testing.T) {
	opts, err := parseOptions("generate list", []string{"--search=Traversal"})
	if err != nil || opts.search != "Traversal" {
		t.Fatalf("options = %+v, error = %v", opts, err)
	}
	if _, err := parseOptions("generate list", []string{"traversal"}); err == nil {
		t.Fatal("accepted a positional argument")
	}
	if _, err := parseOptions("generate list", []string{"--format=json"}); err == nil {
		t.Fatal("accepted an unsupported format option")
	}
	if _, err := parseOptions("generate list", []string{"--help"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("help error = %v, want flag.ErrHelp", err)
	}
}
