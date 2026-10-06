// Package indexcmd defines the command that compiles catalog metadata into the
// distributable index.
package indexcmd

import (
	"path/filepath"

	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/pkg/constants"
)

const (
	commandName    = "index"
	commandSummary = "Compile the master index.json from problems/"
)

type options struct {
	output string
}

func Command() cli.Command {
	return cli.Command{Name: commandName, Summary: commandSummary, Run: cli.Handle(parseOptions, generateIndex)}
}

func Execute(args []string) error {
	return cli.Execute(Command(), args)
}

func parseOptions(commandPath string, args []string) (options, error) {
	var opts options
	parser := cli.NewFlagParser(commandPath, "[options]")
	parser.StringVar(&opts.output, defaultOutputPath(), cli.Option{
		Name:        "output",
		Value:       "path",
		Description: "Path for the compiled index.json",
	})
	if err := parser.Parse(args); err != nil {
		return options{}, err
	}
	return opts, nil
}

func defaultOutputPath() string {
	return filepath.Join(constants.DistDir, "index.json")
}
