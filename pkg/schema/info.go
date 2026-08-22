package schema

import "fmt"

// Difficulty represents the concept difficulty level.
type Difficulty string

const (
	DifficultyEasy   Difficulty = "Easy"
	DifficultyMedium Difficulty = "Medium"
	DifficultyHard   Difficulty = "Hard"
)

// Validate ensures difficulty matches supported values.
func (d Difficulty) Validate() error {
	switch d {
	case DifficultyEasy, DifficultyMedium, DifficultyHard:
		return nil
	default:
		return fmt.Errorf("invalid difficulty %q: must be Easy, Medium, or Hard", d)
	}
}

// ItemRef represents an ordered child entry in a parent info.json.
type ItemRef struct {
	ID    string `json:"id"`
	Order int    `json:"order"`
}

// ParentInfo models root and category/sub-category info.json files.
type ParentInfo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Items       []ItemRef `json:"items"`
}

// LeafInfo models problem-level info.json files.
type LeafInfo struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Difficulty  Difficulty `json:"difficulty"`
	Description string     `json:"description,omitempty"`
}
