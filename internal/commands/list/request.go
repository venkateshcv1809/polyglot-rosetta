package listcmd

import "polyglot-rosetta/internal/cli"

type options struct {
	search string
}

func parseOptions(commandPath string, args []string) (options, error) {
	var opts options
	parser := cli.NewFlagParser(commandPath, "[options]")
	parser.StringVar(&opts.search, "", cli.Option{
		Name: "search", Value: "text", Description: "Filter concept paths and titles",
	})
	if err := parser.Parse(args); err != nil {
		return options{}, err
	}
	return opts, nil
}
