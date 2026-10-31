package probecmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/domain/language"
)

type options struct {
	language string
	format   string
	clear    bool
}

func parseOptions(commandPath string, args []string) (options, error) {
	return parseOptionsWithOutput(commandPath, args, os.Stdout)
}

func parseOptionsWithOutput(commandPath string, args []string, output io.Writer) (options, error) {
	var opts options
	parser := cli.NewFlagParser(commandPath, "[options]")
	parser.StringVar(&opts.language, "", cli.Option{
		Name:        "lang",
		Value:       "language",
		Description: "Language to display; cache refresh checks all supported languages",
		Default:     "all",
	})
	parser.StringVar(&opts.format, "text", cli.Option{
		Name: "format", Value: "text|json", Description: "Output format", Default: "text",
	})
	parser.BoolVar(&opts.clear, false, cli.Option{
		Name: "clear", Description: "Remove cached toolchain results without probing",
	})
	cli.DiscardUsageIfJSON(parser, args, output)
	if err := parser.Parse(args); err != nil {
		return options{}, err
	}
	if strings.EqualFold(strings.TrimSpace(opts.language), "all") {
		opts.language = ""
	} else if opts.language != "" {
		parsed, err := language.Parse(opts.language)
		if err != nil {
			parser.PrintUsage()
			return options{}, fmt.Errorf("unsupported language %q", opts.language)
		}
		opts.language = string(parsed)
	}
	if err := cli.ValidateOutputFormat(opts.format); err != nil {
		parser.PrintUsage()
		return options{}, err
	}
	if opts.clear && opts.language != "" {
		parser.PrintUsage()
		return options{}, fmt.Errorf("--clear cannot be combined with --lang: the cache covers all languages")
	}
	return opts, nil
}
