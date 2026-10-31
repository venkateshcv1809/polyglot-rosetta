package scaffoldcmd

import (
	"testing"

	"polyglot-rosetta/internal/cli"
)

func TestScaffoldSubcommands(t *testing.T) {
	commands := []struct {
		name    string
		summary string
		command cli.Command
	}{
		{name: categoryCommandName, summary: categoryCommandSummary, command: CategoryCommand()},
		{name: conceptCommandName, summary: conceptCommandSummary, command: ConceptCommand()},
	}

	for _, test := range commands {
		t.Run(test.name, func(t *testing.T) {
			if test.command.Name != test.name {
				t.Fatalf("name = %q, want %q", test.command.Name, test.name)
			}
			if test.command.Summary != test.summary {
				t.Fatalf("summary = %q, want %q", test.command.Summary, test.summary)
			}
			if test.command.Run == nil {
				t.Fatal("Run is nil")
			}
		})
	}
}
