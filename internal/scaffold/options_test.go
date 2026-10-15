package scaffold

import (
	"testing"

	"polyglot-rosetta/internal/domain/catalog"
)

func TestCreateCategoryOptionsValidate(t *testing.T) {
	valid := CreateCategoryOptions{ID: "data-structures", Title: "Data Structures"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate valid options: %v", err)
	}
	for _, options := range []CreateCategoryOptions{
		{Title: "Missing ID"},
		{ID: "Invalid ID", Title: "Invalid ID"},
		{ID: "missing-title"},
	} {
		if err := options.Validate(); err == nil {
			t.Errorf("Validate(%+v) returned nil error", options)
		}
	}
}

func TestCreateConceptOptionsValidate(t *testing.T) {
	valid := CreateConceptOptions{ID: "hello-world", Title: "Hello World", Difficulty: catalog.DifficultyEasy}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate valid options: %v", err)
	}
	for _, options := range []CreateConceptOptions{
		{Title: "Missing ID", Difficulty: catalog.DifficultyEasy},
		{ID: "Invalid ID", Title: "Invalid ID", Difficulty: catalog.DifficultyEasy},
		{ID: "missing-title", Difficulty: catalog.DifficultyEasy},
		{ID: "bad-difficulty", Title: "Bad Difficulty", Difficulty: "Impossible"},
	} {
		if err := options.Validate(); err == nil {
			t.Errorf("Validate(%+v) returned nil error", options)
		}
	}
}
