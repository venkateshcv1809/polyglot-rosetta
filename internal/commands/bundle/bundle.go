package bundlecmd

import (
	"io"
	"os"

	catalogbundle "polyglot-rosetta/internal/catalog/bundle"
	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/config"
)

func execute(opts options) error {
	return bundleFrom(config.CatalogDir, config.TemplateDir, opts.output, os.Stdout)
}

func bundleFrom(problemsDir, templatesDir, outputDir string, output io.Writer) error {
	if err := catalogbundle.Build(problemsDir, templatesDir, outputDir); err != nil {
		return cli.CommandFailure("Bundle not created", err)
	}
	return printCreatedBundle(output, outputDir)
}
