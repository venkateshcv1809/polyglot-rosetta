package scaffoldcmd

import (
	"fmt"
	"os"

	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/pkg/constants"
	"polyglot-rosetta/pkg/scaffold"
	"polyglot-rosetta/pkg/schema"
)

type conceptOptions struct {
	parent      string
	id          string
	title       string
	description string
	difficulty  string
}

func runConcept(opts conceptOptions) error {
	prompted, err := promptConceptOptions(opts)
	if err != nil {
		return err
	}
	opts = prompted

	scaffolder := scaffold.New(constants.ProblemsDir, constants.TemplateDir)
	createOpts := scaffold.CreateConceptOptions{
		ParentPath:  opts.parent,
		ID:          opts.id,
		Title:       opts.title,
		Description: opts.description,
		Difficulty:  schema.Difficulty(opts.difficulty),
	}
	if err := scaffolder.CreateConcept(createOpts); err != nil {
		return fmt.Errorf("failed to scaffold concept: %w", err)
	}
	fmt.Printf("scaffolded concept %q under %q\n", createOpts.ID, createOpts.ParentPath)
	return nil
}

func parseConceptOptions(commandPath string, args []string) (conceptOptions, error) {
	var opts conceptOptions
	parser := cli.NewFlagParser(commandPath, "[options]")
	parser.StringVar(&opts.parent, "", cli.Option{Name: "parent", Value: "path", Description: "Parent category path; prompted when omitted"})
	parser.StringVar(&opts.id, "", cli.Option{Name: "id", Value: "id", Description: "Concept identifier; prompted when omitted"})
	parser.StringVar(&opts.title, "", cli.Option{Name: "title", Value: "title", Description: "Concept title; prompted when omitted"})
	parser.StringVar(&opts.description, "", cli.Option{Name: "description", Value: "text", Description: "Concept description; prompted when omitted"})
	parser.StringVar(&opts.difficulty, "", cli.Option{Name: "difficulty", Value: "level", Description: "Difficulty: Easy, Medium, or Hard", Default: constants.ConceptDifficulty})

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
