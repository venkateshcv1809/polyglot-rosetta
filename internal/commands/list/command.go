// Package listcmd defines the local catalog discovery command.
package listcmd

import "polyglot-rosetta/internal/cli"

const (
	commandName    = "list"
	commandSummary = "List concept paths and run IDs from problems/"
)

// Command returns the list command definition for the generate binary.
func Command() cli.Command {
	return cli.Command{Name: commandName, Summary: commandSummary, Run: cli.Handle(parseOptions, list)}
}

// Execute parses args and runs the list command directly.
func Execute(args []string) error {
	return cli.Execute(Command(), args)
}
