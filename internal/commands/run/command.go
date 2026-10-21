// Package runcmd defines the command for executing source code against catalog
// test cases.
package runcmd

import "polyglot-rosetta/internal/cli"

const (
	commandName    = "run"
	commandSummary = "Execute code against test cases"
)

// Command returns the run command definition for use by any CLI entrypoint.
func Command() cli.Command {
	return cli.Command{Name: commandName, Summary: commandSummary, Run: runCommand}
}
