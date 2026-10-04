package scaffoldcmd

import (
	"fmt"
	"os"

	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/pkg/constants"
	"polyglot-rosetta/pkg/scaffold"
)

type categoryOptions struct {
	id          string
	title       string
	description string
	parent      string
}

func runCategory(opts categoryOptions) error {
	prompted, err := promptCategoryOptions(opts)
	if err != nil {
		return err
	}
	opts = prompted

	scaffolder := scaffold.New(constants.ProblemsDir, constants.TemplateDir)
	createOpts := scaffold.CreateCategoryOptions{
		ID:          opts.id,
		Title:       opts.title,
		Description: opts.description,
		ParentPath:  opts.parent,
	}
	if err := scaffolder.CreateCategory(createOpts); err != nil {
		return fmt.Errorf("failed to scaffold category: %w", err)
	}
	fmt.Printf("scaffolded category %q\n", createOpts.ID)
	return nil
}

func parseCategoryOptions(commandPath string, args []string) (categoryOptions, error) {
	var opts categoryOptions
	parser := cli.NewFlagParser(commandPath, "[options]")
	parser.StringVar(&opts.id, "", cli.Option{Name: "id", Value: "id", Description: "Category identifier; prompted when omitted"})
	parser.StringVar(&opts.title, "", cli.Option{Name: "title", Value: "title", Description: "Category title; prompted when omitted"})
	parser.StringVar(&opts.description, "", cli.Option{Name: "description", Value: "text", Description: "Category description; prompted when omitted"})
	parser.StringVar(&opts.parent, "", cli.Option{Name: "parent", Value: "path", Description: "Parent category path; prompted when omitted"})

	if err := parser.Parse(args); err != nil {
		return categoryOptions{}, err
	}
	return opts, nil
}

func promptCategoryOptions(opts categoryOptions) (categoryOptions, error) {
	answers, err := cli.PromptForm(os.Stdin, os.Stdout, CategoryFormSpec, map[string]string{
		"id":          opts.id,
		"title":       opts.title,
		"description": opts.description,
		"parent":      opts.parent,
	})
	if err != nil {
		return categoryOptions{}, fmt.Errorf("collect category options: %w", err)
	}
	return categoryOptions{
		id:          answers["id"],
		title:       answers["title"],
		description: answers["description"],
		parent:      answers["parent"],
	}, nil
}
