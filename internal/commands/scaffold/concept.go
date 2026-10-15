package scaffoldcmd

import (
	"os"

	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/config"
	"polyglot-rosetta/internal/domain/catalog"
	"polyglot-rosetta/internal/scaffold"
)

func runConcept(opts conceptOptions) error {
	prompted, err := promptConceptOptions(opts)
	if err != nil {
		return err
	}
	opts = prompted

	scaffolder := scaffold.New(config.CatalogDir, config.TemplateDir)
	createOpts := scaffold.CreateConceptOptions{
		ParentPath:  opts.parent,
		ID:          opts.id,
		Title:       opts.title,
		Description: opts.description,
		Difficulty:  catalog.Difficulty(opts.difficulty),
	}
	if err := scaffolder.CreateConcept(createOpts); err != nil {
		return cli.CommandFailure("Concept not created", err)
	}
	return printCreatedConcept(os.Stdout, createOpts)
}
