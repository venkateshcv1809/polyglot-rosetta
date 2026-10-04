// Package probecmd defines the command for inspecting locally available
// language toolchains.
package probecmd

import "polyglot-rosetta/internal/cli"

const (
	commandName    = "probe"
	commandSummary = "Inspect local language toolchains"
)

type options struct {
	language string
}

func Command() cli.Command {
	return cli.Command{Name: commandName, Summary: commandSummary, Run: cli.Handle(parseOptions, probe)}
}

func Execute(args []string) error {
	return cli.Execute(Command(), args)
}

func parseOptions(commandPath string, args []string) (options, error) {
	var opts options
	parser := cli.NewFlagParser(commandPath, "[options]")
	parser.StringVar(&opts.language, "", cli.Option{
		Name:        "lang",
		Value:       "language",
		Description: "Language to probe: go, python, rust, typescript, or zig",
		Default:     "all",
	})
	if err := parser.Parse(args); err != nil {
		return options{}, err
	}
	return opts, nil
}
