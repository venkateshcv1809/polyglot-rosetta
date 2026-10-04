package generatecmd

import (
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestCommand(t *testing.T) {
	command := Command()
	if command.Name != commandName || command.Summary != commandSummary {
		t.Fatalf("command = %#v", command)
	}
	wantNames := []string{"category", "concept", "index"}
	if len(command.Subcommands) != len(wantNames) {
		t.Fatalf("subcommands = %d, want %d", len(command.Subcommands), len(wantNames))
	}
	for i, want := range wantNames {
		if got := command.Subcommands[i].Name; got != want {
			t.Fatalf("subcommand %d = %q, want %q", i, got, want)
		}
	}
}

func TestExecute_UnknownCommand(t *testing.T) {
	err := Execute([]string{"manifest"})
	if err == nil {
		t.Fatal("expected unknown command error")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("error = %v", err)
	}
}

func TestExecute_MissingCommand(t *testing.T) {
	if err := Execute(nil); err == nil {
		t.Fatal("expected missing command error")
	}
}

func TestExecute_Help(t *testing.T) {
	if err := Execute([]string{"-h"}); err != nil {
		t.Fatalf("help: %v", err)
	}
}

func TestExecute_CategoryHelpForwardsSubcommand(t *testing.T) {
	err := Execute([]string{"category", "-h"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("got %v, want flag.ErrHelp (category must be forwarded to scaffold)", err)
	}
}

func TestExecute_IndexUnknownFlag(t *testing.T) {
	if err := Execute([]string{"index", "-bogus"}); err == nil {
		t.Fatal("expected flag parse error")
	}
}
