package scaffoldcmd

import (
	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/catalog"
)

// CategoryFormSpec defines the prompt order and rules for creating categories.
var CategoryFormSpec = []cli.PromptField{
	{
		Key:      "id",
		Prompt:   "Category directory name (e.g. 'sorting')",
		Required: true,
	},
	{
		Key:      "title",
		Prompt:   "Category Title (e.g. 'Sorting Algorithms')",
		Required: true,
	},
	{
		Key:      "description",
		Prompt:   "Category Description",
		Required: false,
	},
	{
		Key:      "parent",
		Prompt:   "Parent Category Path (e.g. 'data-structures')",
		Required: false,
	},
}

// ConceptFormSpec defines the prompt order and rules for creating concepts.
var ConceptFormSpec = []cli.PromptField{
	{
		Key:      "parent",
		Prompt:   "Parent Category Path (e.g. 'algorithms/sorting')",
		Required: true,
	},
	{
		Key:      "id",
		Prompt:   "Concept directory name (e.g. 'two-sum')",
		Required: true,
	},
	{
		Key:      "title",
		Prompt:   "Concept Title (e.g. 'Two Sum')",
		Required: true,
	},
	{
		Key:      "description",
		Prompt:   "Concept Description",
		Required: false,
	},
	{
		Key:          "difficulty",
		Prompt:       "Difficulty Rating (Easy, Medium, Hard)",
		Required:     false,
		DefaultValue: string(catalog.DifficultyEasy),
	},
}
