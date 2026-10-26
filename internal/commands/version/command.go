// Package versioncmd prints the public CLI version.
package versioncmd

import "polyglot-rosetta/internal/cli"

const (
	commandName    = "version"
	commandSummary = "Print the CLI version"
)

// Command returns the public version command.
func Command() cli.Command {
	return cli.Command{Name: commandName, Summary: commandSummary, Run: versionCommand}
}
