// Package runcmd defines the command for executing source code against catalog
// test cases.
package runcmd

import (
	"fmt"

	"polyglot-rosetta/internal/cli"
)

const (
	commandName     = "run"
	commandSummary  = "Execute code against test cases"
	defaultLanguage = "go"
)

type options struct {
	concept  string
	language string
	all      bool
	path     string
	stdin    bool
	code     string
}

func Command() cli.Command {
	return cli.Command{Name: commandName, Summary: commandSummary, Run: cli.Handle(parseOptions, run)}
}

func Execute(args []string) error {
	return cli.Execute(Command(), args)
}

func parseOptions(commandPath string, args []string) (options, error) {
	var opts options
	parser := cli.NewFlagParser(commandPath,
		"--concept=<id> --path=<file> [options]",
		"--concept=<id> --stdin [options]",
		"--concept=<id> --code=<string> [options]",
	)
	parser.StringVar(&opts.concept, "", cli.Option{Name: "concept", Value: "id", Description: "Target concept identifier", Required: true})
	parser.StringVar(&opts.path, "", cli.Option{Name: "path", Value: "file", Description: "Read source code from a file; one input source is required"})
	parser.BoolVar(&opts.stdin, false, cli.Option{Name: "stdin", Description: "Read source code from standard input; one input source is required"})
	parser.StringVar(&opts.code, "", cli.Option{Name: "code", Value: "string", Description: "Use inline source code; one input source is required"})
	parser.StringVar(&opts.language, defaultLanguage, cli.Option{Name: "lang", Value: "language", Description: "Target programming language"})
	parser.BoolVar(&opts.all, false, cli.Option{Name: "all", Description: "Include hidden test cases"})

	if err := parser.Parse(args); err != nil {
		return options{}, err
	}
	if opts.concept == "" {
		parser.PrintUsage()
		return options{}, fmt.Errorf("missing required option: --concept")
	}
	if opts.path == "" && !opts.stdin && opts.code == "" {
		parser.PrintUsage()
		return options{}, fmt.Errorf("missing input source: provide one of --path, --stdin, or --code")
	}
	return opts, nil
}
