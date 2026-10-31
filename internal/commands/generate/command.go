// Package generatecmd defines the maintainer command group for scaffolding,
// listing, and packaging catalog content.
package generatecmd

import (
	"polyglot-rosetta/internal/cli"
	bundlecmd "polyglot-rosetta/internal/commands/bundle"
	listcmd "polyglot-rosetta/internal/commands/list"
	scaffoldcmd "polyglot-rosetta/internal/commands/scaffold"
)

const (
	commandName    = "generate"
	commandSummary = "Scaffold, list, and package local catalog content."
)

func Command() cli.Command {
	return cli.Command{
		Name:    commandName,
		Summary: commandSummary,
		Subcommands: []cli.Command{
			bundlecmd.Command(),
			scaffoldcmd.CategoryCommand(),
			scaffoldcmd.ConceptCommand(),
			listcmd.Command(),
		},
	}
}

func Execute(args []string) error {
	return cli.Execute(Command(), args)
}
