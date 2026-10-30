package bundlecmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/config"
)

type options struct {
	output string
}

func parseOptions(commandPath string, args []string) (options, error) {
	var opts options
	parser := cli.NewFlagParser(commandPath, "[options]")
	parser.StringVar(&opts.output, filepath.Join(config.DistributionDir, config.CatalogDirName), cli.Option{
		Name: "output", Value: "path", Description: "Directory for the problem archives and index",
	})
	if err := parser.Parse(args); err != nil {
		return options{}, err
	}
	if strings.TrimSpace(opts.output) == "" {
		parser.PrintUsage()
		return options{}, fmt.Errorf("output directory cannot be empty")
	}
	return opts, nil
}
