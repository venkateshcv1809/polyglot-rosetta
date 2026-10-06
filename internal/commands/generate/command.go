// Package generatecmd defines the maintainer command group for scaffolding
// catalog content and compiling the generated index.
package generatecmd

import (
	"polyglot-rosetta/internal/cli"
	indexcmd "polyglot-rosetta/internal/commands/indexer"
	scaffoldcmd "polyglot-rosetta/internal/commands/scaffold"
)

const (
	commandName    = "generate"
	commandSummary = "Scaffold catalog content and compile index.json."
)

func Command() cli.Command {
	return cli.Command{
		Name:    commandName,
		Summary: commandSummary,
		Subcommands: []cli.Command{
			scaffoldcmd.CategoryCommand(),
			scaffoldcmd.ConceptCommand(),
			indexcmd.Command(),
		},
	}
}

func Execute(args []string) error {
	return cli.Execute(Command(), args)
}
