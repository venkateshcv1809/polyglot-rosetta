// Package scaffoldcmd defines the category and concept scaffolding commands used
// by the maintainer generate command.
package scaffoldcmd

import "polyglot-rosetta/internal/cli"

const (
	commandName            = "generate"
	commandSummary         = "Scaffold catalog content."
	categoryCommandName    = "category"
	categoryCommandSummary = "Scaffold a category directory and info.json"
	conceptCommandName     = "concept"
	conceptCommandSummary  = "Scaffold a concept directory and language templates"
)

func Command() cli.Command {
	return cli.Command{
		Name:        commandName,
		Summary:     commandSummary,
		Subcommands: []cli.Command{CategoryCommand(), ConceptCommand()},
	}
}

func Execute(args []string) error {
	return cli.Execute(Command(), args)
}

func CategoryCommand() cli.Command {
	return cli.Command{
		Name:    categoryCommandName,
		Summary: categoryCommandSummary,
		Run:     cli.Handle(parseCategoryOptions, runCategory),
	}
}

func ConceptCommand() cli.Command {
	return cli.Command{
		Name:    conceptCommandName,
		Summary: conceptCommandSummary,
		Run:     cli.Handle(parseConceptOptions, runConcept),
	}
}
