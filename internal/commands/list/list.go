package listcmd

import (
	"io"
	"os"

	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/config"
)

func list(opts options) error {
	return listFrom(config.CatalogDir, opts.search, os.Stdout)
}

func listFrom(root, search string, output io.Writer) error {
	metadata, err := catalog.Build(root)
	if err != nil {
		return cli.CommandFailure("Concepts not listed", err)
	}
	return writeList(output, filterConcepts(orderedConcepts(metadata), search), search)
}
