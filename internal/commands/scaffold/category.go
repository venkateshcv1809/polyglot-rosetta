package scaffoldcmd

import (
	"os"

	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/config"
	"polyglot-rosetta/internal/scaffold"
)

func runCategory(opts categoryOptions) error {
	prompted, err := promptCategoryOptions(opts)
	if err != nil {
		return err
	}
	opts = prompted

	scaffolder := scaffold.New(config.CatalogDir, config.TemplateDir)
	createOpts := scaffold.CreateCategoryOptions{
		ID:          opts.id,
		Title:       opts.title,
		Description: opts.description,
		ParentPath:  opts.parent,
	}
	if err := scaffolder.CreateCategory(createOpts); err != nil {
		return cli.CommandFailure("Category not created", err)
	}
	return printCreatedCategory(os.Stdout, createOpts)
}
