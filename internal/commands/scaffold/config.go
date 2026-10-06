package scaffoldcmd

import (
	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/pkg/constants"
)

// Prompt order and rules for creating categories.
var CategoryFormSpec = []cli.PromptField{
	{
		Key:      "id",
		Prompt:   "Category ID slug (e.g. 'sorting')",
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

// Prompt order and rules for creating concepts.
var ConceptFormSpec = []cli.PromptField{
	{
		Key:      "parent",
		Prompt:   "Parent Category Path (e.g. 'algorithms/sorting')",
		Required: true,
	},
	{
		Key:      "id",
		Prompt:   "Concept ID slug (e.g. 'two-sum')",
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
		DefaultValue: constants.ConceptDifficulty,
	},
}
