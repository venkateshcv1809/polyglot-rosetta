package scaffoldcmd

import (
	"flag"
	"fmt"

	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/pkg/constants"
	"polyglot-rosetta/pkg/scaffold"
	"polyglot-rosetta/pkg/schema"
)

// Handle command-line arguments for scaffold operations.
func Execute(args []string) error {
	if len(args) < 1 {
		PrintUsage()
		return fmt.Errorf("missing subcommand: expected 'category' or 'concept'")
	}

	subcommand := args[0]
	subArgs := args[1:]

	scaffolder := scaffold.New(constants.SrcDir, constants.TemplateDir)

	switch subcommand {
	case "category":
		return runCategory(scaffolder, subArgs)
	case "concept":
		return runConcept(scaffolder, subArgs)
	default:
		PrintUsage()
		return fmt.Errorf("unknown scaffold subcommand %q: expected 'category' or 'concept'", subcommand)
	}
}

func PrintUsage() {
	fmt.Println("Scaffold Commands:")
	fmt.Println("  category [--id=<id>] [--title=<title>] [--description=<desc>] [--parent=<path>]")
	fmt.Println("  concept  [--id=<id>] [--title=<title>] [--parent=<path>] [--difficulty=<Easy|Medium|Hard>] [--description=<desc>]")
}

func runCategory(s *scaffold.Scaffolder, args []string) error {
	fs := flag.NewFlagSet("scaffold category", flag.ExitOnError)

	var id, title, description, parent string
	fs.StringVar(&id, "id", "", "Category ID")
	fs.StringVar(&title, "title", "", "Category Title")
	fs.StringVar(&description, "description", "", "Category Description")
	fs.StringVar(&parent, "parent", "", "Parent Path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	answers := cli.PromptForm(CategoryFormSpec, map[string]string{
		"id":          id,
		"title":       title,
		"description": description,
		"parent":      parent,
	})

	opts := scaffold.CreateCategoryOptions{
		ID:          answers["id"],
		Title:       answers["title"],
		Description: answers["description"],
		ParentPath:  answers["parent"],
	}

	if err := s.CreateCategory(opts); err != nil {
		return fmt.Errorf("failed to scaffold category: %w", err)
	}

	fmt.Printf("✅ Successfully scaffolded category %q\n", opts.ID)
	return nil
}

func runConcept(s *scaffold.Scaffolder, args []string) error {
	fs := flag.NewFlagSet("scaffold concept", flag.ExitOnError)

	var id, title, description, parent, difficulty string
	fs.StringVar(&id, "id", "", "Concept ID")
	fs.StringVar(&title, "title", "", "Concept Title")
	fs.StringVar(&description, "description", "", "Concept Description")
	fs.StringVar(&parent, "parent", "", "Parent Path")
	fs.StringVar(&difficulty, "difficulty", "", "Difficulty (Easy, Medium, Hard)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	answers := cli.PromptForm(ConceptFormSpec, map[string]string{
		"id":          id,
		"title":       title,
		"description": description,
		"parent":      parent,
		"difficulty":  difficulty,
	})

	opts := scaffold.CreateConceptOptions{
		ID:          answers["id"],
		Title:       answers["title"],
		Description: answers["description"],
		ParentPath:  answers["parent"],
		Difficulty:  schema.Difficulty(answers["difficulty"]),
	}

	if err := s.CreateConcept(opts); err != nil {
		return fmt.Errorf("failed to scaffold concept: %w", err)
	}

	fmt.Printf("✅ Successfully scaffolded concept %q under %q\n", opts.ID, opts.ParentPath)
	return nil
}
