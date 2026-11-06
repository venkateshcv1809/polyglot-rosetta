package sitecmd

import (
	"path/filepath"

	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/config"
)

const defaultPagesRoot = ".github/pages"

type options struct {
	source  string
	output  string
	catalog string
}

func parseSiteOptions(commandPath string, args []string) (options, error) {
	opts := options{output: defaultPagesRoot}
	parser := cli.NewFlagParser(commandPath, "[options]")
	bindSource(parser, &opts.source)
	bindCatalog(parser, &opts.catalog)
	if err := parser.Parse(args); err != nil {
		return options{}, err
	}
	return opts, nil
}

func bindSource(parser *cli.FlagParser, target *string) {
	parser.StringVar(target, "", cli.Option{
		Name:        "source",
		Value:       "path",
		Description: "Local portfolio checkout directory to copy from. When omitted, PORTFOLIO_ROOT or a shallow clone is used",
	})
}

func bindCatalog(parser *cli.FlagParser, target *string) {
	parser.StringVar(target, filepath.Join(config.DistributionDir, config.CatalogDirName), cli.Option{
		Name: "catalog", Value: "path", Description: "Packaged catalog directory copied into the site",
	})
}
