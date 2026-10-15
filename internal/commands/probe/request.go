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
	if requestsJSON(args) {
		parser.SetOutput(io.Discard)
	} else {
		parser.SetOutput(output)
	}
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
	if opts.format != "text" && opts.format != "json" {
		parser.PrintUsage()
		return options{}, fmt.Errorf("unsupported output format %q: use text or json", opts.format)
	}
	if opts.clear && opts.language != "" {
		parser.PrintUsage()
		return options{}, fmt.Errorf("--clear cannot be combined with --lang: the cache covers all languages")
	}
	return opts, nil
}

// Detect JSON before parsing so malformed flags can still receive a JSON error.
func requestsJSON(args []string) bool {
	format := ""
	for i, arg := range args {
		if cli.IsHelp(arg) {
			return false
		}
		if arg == "--format" || arg == "-format" {
			if i+1 < len(args) {
				format = args[i+1]
			}
			continue
		}
		if value, found := strings.CutPrefix(arg, "--format="); found {
			format = value
		} else if value, found := strings.CutPrefix(arg, "-format="); found {
			format = value
		}
	}
	return format == "json"
}
