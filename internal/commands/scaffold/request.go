package scaffoldcmd

import (
	"fmt"
	"os"

	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/domain/catalog"
)

type categoryOptions struct {
	id          string
	title       string
	description string
	parent      string
}

type conceptOptions struct {
	parent      string
	id          string
	title       string
	description string
	difficulty  string
}

func parseCategoryOptions(commandPath string, args []string) (categoryOptions, error) {
	var opts categoryOptions
	parser := cli.NewFlagParser(commandPath, "[options]")
	parser.StringVar(&opts.id, "", cli.Option{Name: "id", Value: "name", Description: "Category directory name under the parent; prompted when omitted"})
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

func parseConceptOptions(commandPath string, args []string) (conceptOptions, error) {
	var opts conceptOptions
	parser := cli.NewFlagParser(commandPath, "[options]")
	parser.StringVar(&opts.parent, "", cli.Option{Name: "parent", Value: "path", Description: "Parent category path; prompted when omitted"})
	parser.StringVar(&opts.id, "", cli.Option{Name: "id", Value: "name", Description: "Concept directory name under the parent; prompted when omitted"})
	parser.StringVar(&opts.title, "", cli.Option{Name: "title", Value: "title", Description: "Concept title; prompted when omitted"})
	parser.StringVar(&opts.description, "", cli.Option{Name: "description", Value: "text", Description: "Concept description; prompted when omitted"})
	parser.StringVar(&opts.difficulty, "", cli.Option{Name: "difficulty", Value: "level", Description: "Difficulty: Easy, Medium, or Hard", Default: string(catalog.DifficultyEasy)})

	if err := parser.Parse(args); err != nil {
		return conceptOptions{}, err
	}
	return opts, nil
}

func promptConceptOptions(opts conceptOptions) (conceptOptions, error) {
	answers, err := cli.PromptForm(os.Stdin, os.Stdout, ConceptFormSpec, map[string]string{
		"parent":      opts.parent,
		"id":          opts.id,
		"title":       opts.title,
		"description": opts.description,
		"difficulty":  opts.difficulty,
	})
	if err != nil {
		return conceptOptions{}, fmt.Errorf("collect concept options: %w", err)
	}
	return conceptOptions{
		parent:      answers["parent"],
		id:          answers["id"],
		title:       answers["title"],
		description: answers["description"],
		difficulty:  answers["difficulty"],
	}, nil
}
