package catalogcmd

import (
	"io"
	"os"
	"strings"

	catalogsync "polyglot-rosetta/internal/catalog/sync"
	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/config"
	"polyglot-rosetta/internal/version"
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
	parser := cli.NewFlagParser(commandPath, "[options]")
	localCatalog := strings.TrimSpace(os.Getenv(config.CatalogEnvironment)) != ""
	defaultURL := strings.TrimSpace(os.Getenv(config.CatalogURLEnvironment))
	if defaultURL == "" && !version.IsDevelopment() {
		defaultURL = catalogsync.DefaultURL
	}
	helpDefault := "local problems/"
	if !version.IsDevelopment() {
		helpDefault = catalogsync.DefaultURL
	}
	if localCatalog {
		defaultURL = ""
		helpDefault = "local PROSETTA_CATALOG"
	}
	parser.StringVar(&opts.url, defaultURL, cli.Option{
		Name: "url", Value: "directory-url", Description: "Static URL containing manifest.json and content archives; ignored when PROSETTA_CATALOG is set", Default: helpDefault,
	})
	parser.StringVar(&opts.format, "text", cli.Option{
		Name: "format", Value: "text|json", Description: "Output format", Default: "text",
	})
	cli.DiscardUsageIfJSON(parser, args, output)
	if err := parser.Parse(args); err != nil {
		return options{}, err
	}
	if localCatalog {
		opts.url = ""
	} else if strings.TrimSpace(opts.url) == "" && !version.IsDevelopment() {
		opts.url = catalogsync.DefaultURL
	}
	if err := cli.ValidateOutputFormat(opts.format); err != nil {
		parser.PrintUsage()
		return options{}, err
	}
	cache, err := catalogsync.DefaultDir()
	if err != nil {
		return options{}, err
	}
	opts.cache = cache
	return opts, nil
}
