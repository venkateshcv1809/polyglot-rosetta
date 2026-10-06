package scaffoldcmd

import (
	"testing"

	"polyglot-rosetta/internal/cli"
)

func TestExecuteHelp(t *testing.T) {
	if err := Execute([]string{"--help"}); err != nil {
		t.Fatalf("help returned an error: %v", err)
	}
}

func TestCommand(t *testing.T) {
	command := Command()
	if command.Name != commandName || command.Summary != commandSummary {
		t.Fatalf("command = %#v", command)
	}
	if len(command.Subcommands) != 2 {
		t.Fatalf("subcommands = %d, want 2", len(command.Subcommands))
	}
	if command.Subcommands[0].Name != categoryCommandName || command.Subcommands[1].Name != conceptCommandName {
		t.Fatalf("subcommand order = [%q, %q]", command.Subcommands[0].Name, command.Subcommands[1].Name)
	}
}

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
