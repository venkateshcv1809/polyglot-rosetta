package catalogcmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	catalogsync "polyglot-rosetta/internal/catalog/sync"
	"polyglot-rosetta/internal/cli"
)

type options struct {
	url    string
	cache  string
	format string
}

func parseOptions(commandPath string, args []string) (options, error) {
	return parseOptionsWithOutput(commandPath, args, os.Stdout)
}

func parseOptionsWithOutput(commandPath string, args []string, output io.Writer) (options, error) {
	var opts options
	parser := cli.NewFlagParser(commandPath, "--url=<directory-url> [options]")
	parser.StringVar(&opts.url, os.Getenv("PROSETTA_CATALOG_URL"), cli.Option{
		Name: "url", Value: "directory-url", Description: "Static URL containing manifest.json and content archives", Required: true,
	})
	parser.StringVar(&opts.format, "text", cli.Option{
		Name: "format", Value: "text|json", Description: "Output format", Default: "text",
	})
	if requestsJSON(args) {
		parser.SetOutput(io.Discard)
	} else {
		parser.SetOutput(output)
	}
	if err := parser.Parse(args); err != nil {
		return options{}, err
	}
	if strings.TrimSpace(opts.url) == "" {
		parser.PrintUsage()
		return options{}, fmt.Errorf("missing required option: --url")
	}
	if opts.format != "text" && opts.format != "json" {
		parser.PrintUsage()
		return options{}, fmt.Errorf("unsupported output format %q: use text or json", opts.format)
	}
	cache, err := catalogsync.DefaultDir()
	if err != nil {
		return options{}, err
	}
	opts.cache = cache
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
