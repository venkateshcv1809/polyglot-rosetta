// Package sitecmd defines the site binary that stages the shared Pages theme.
package sitecmd

import "polyglot-rosetta/internal/cli"

const (
	commandName    = "site"
	commandSummary = "Stage the shared GitHub Pages theme and preview it locally"
	stageName      = "stage"
	stageSummary   = "Copy theme and catalog files into a Jekyll source directory"
	devName        = "dev"
	devSummary     = "Stage the site and serve it with live reload"
)

// Command returns the root command for the site binary.
func Command() cli.Command {
	return cli.Command{
		Name:    commandName,
		Summary: commandSummary,
		Subcommands: []cli.Command{
			StageCommand(),
			DevCommand(),
		},
	}
}

// StageCommand returns the site stage subcommand.
func StageCommand() cli.Command {
	return cli.Command{Name: stageName, Summary: stageSummary, Run: cli.Handle(parseSiteOptions, runStage)}
}

// DevCommand returns the site dev subcommand.
func DevCommand() cli.Command {
	return cli.Command{Name: devName, Summary: devSummary, Run: cli.Handle(parseSiteOptions, runDev)}
}

// Execute runs the site command tree.
func Execute(args []string) error {
	return cli.Execute(Command(), args)
}
