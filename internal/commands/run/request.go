package runcmd

import (
	"fmt"
	"io"
	"os"

	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/domain/language"
)

// options holds native CLI inputs; the transport-neutral request lives in internal/api/run.
type options struct {
	id       string
	language string
	format   string
	all      bool
	stdin    bool
	code     string
}

func parseOptions(commandPath string, args []string) (options, error) {
	return parseOptionsWithOutput(commandPath, args, os.Stdout)
}

func parseOptionsWithOutput(commandPath string, args []string, output io.Writer) (options, error) {
	var opts options
	jsonOutput := cli.RequestsJSON(args)
	parser := cli.NewFlagParser(commandPath,
		"--id=<id> --lang=<language> [options]",
		"--id=<id> --lang=<language> --stdin [options]",
		"--id=<id> --lang=<language> --code=<string> [options]",
	)
	parser.StringVar(&opts.id, "", cli.Option{Name: "id", Value: "id", Description: "Concept ID or unique prefix from the catalog", Required: true})
	parser.BoolVar(&opts.stdin, false, cli.Option{Name: "stdin", Description: "Read source code from standard input instead of the catalog solution"})
	parser.StringVar(&opts.code, "", cli.Option{Name: "code", Value: "string", Description: "Use inline source code instead of the catalog solution"})
	parser.StringVar(&opts.language, "", cli.Option{Name: "lang", Value: "language", Description: "Target programming language", Required: true})
	parser.StringVar(&opts.format, "text", cli.Option{Name: "format", Value: "text|json", Description: "Output format", Default: "text"})
	parser.BoolVar(&opts.all, false, cli.Option{Name: "all", Description: "Include hidden test cases"})
	cli.DiscardUsageIfJSON(parser, args, output)

	if err := parser.Parse(args); err != nil {
		return opts, err
	}
	if opts.id == "" {
		if !jsonOutput {
			parser.PrintUsage()
		}
		return opts, fmt.Errorf("missing required option: --id")
	}
	if _, err := catalog.NormalizeConceptID(opts.id); err != nil {
		if !jsonOutput {
			parser.PrintUsage()
		}
		return opts, err
	}
	if opts.language == "" {
		if !jsonOutput {
			parser.PrintUsage()
		}
		return opts, fmt.Errorf("missing required option: --lang")
	}
	if opts.stdin && opts.code != "" {
		if !jsonOutput {
			parser.PrintUsage()
		}
		return opts, fmt.Errorf("multiple input sources: provide either --stdin or --code")
	}
	if _, err := language.Parse(opts.language); err != nil {
		if !jsonOutput {
			parser.PrintUsage()
		}
		return opts, fmt.Errorf("unsupported language %q", opts.language)
	}
	if err := cli.ValidateOutputFormat(opts.format); err != nil {
		parser.PrintUsage()
		return opts, err
	}
	return opts, nil
}
