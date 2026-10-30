package versioncmd

import (
	"io"

	"polyglot-rosetta/internal/cli"
)

type options struct {
	format string
}

func parseOptions(commandPath string, args []string, output io.Writer) (options, error) {
	var opts options
	parser := cli.NewFlagParser(commandPath, "[options]")
	parser.StringVar(&opts.format, "text", cli.Option{
		Name: "format", Value: "text|json", Description: "Output format", Default: "text",
	})
	cli.DiscardUsageIfJSON(parser, args, output)
	if err := parser.Parse(args); err != nil {
		return options{}, err
	}
	if err := cli.ValidateOutputFormat(opts.format); err != nil {
		parser.PrintUsage()
		return options{}, err
	}
	return opts, nil
}
