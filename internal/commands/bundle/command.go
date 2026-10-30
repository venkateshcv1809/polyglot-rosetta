// Package bundlecmd defines the maintainer command for packaging catalog assets.
package bundlecmd

import "polyglot-rosetta/internal/cli"

const (
	commandName    = "bundle"
	commandSummary = "Package templates and root categories for distribution"
)

// Command returns the bundle subcommand definition.
func Command() cli.Command {
	return cli.Command{Name: commandName, Summary: commandSummary, Run: cli.Handle(parseOptions, execute)}
}
