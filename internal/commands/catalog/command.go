// Package catalogcmd defines the public catalog synchronization command.
package catalogcmd

import (
	"polyglot-rosetta/internal/cli"
)

const (
	commandName    = "catalog"
	commandSummary = "Download and activate problem content"
)

// Command returns the catalog sync command for the public CLI.
func Command() cli.Command {
	return cli.Command{Name: commandName, Summary: commandSummary, Run: catalogCommand}
}
