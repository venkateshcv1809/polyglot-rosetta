// Package catalog owns the canonical catalog model, concept IDs, and local
// metadata scanner. Packaging and synchronization live in subpackages.
package catalog

import (
	"fmt"
	"regexp"
)

var nodeIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ValidNodeID reports whether id is a lowercase kebab-case catalog directory name.
func ValidNodeID(id string) bool {
	return nodeIDPattern.MatchString(id)
}

// Difficulty represents a concept difficulty level.
type Difficulty string

const (
	DifficultyEasy   Difficulty = "Easy"
	DifficultyMedium Difficulty = "Medium"
	DifficultyHard   Difficulty = "Hard"
)

// Validate reports whether the difficulty is supported.
func (d Difficulty) Validate() error {
	switch d {
	case DifficultyEasy, DifficultyMedium, DifficultyHard:
		return nil
	default:
		return fmt.Errorf("invalid difficulty %q: must be Easy, Medium, or Hard", d)
	}
}

// ItemRef identifies an ordered child in parent catalog metadata.
type ItemRef struct {
	ID    string `json:"id"`
	Order int    `json:"order"`
}

// ParentInfo models root, category, and subcategory metadata.
type ParentInfo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Items       []ItemRef `json:"items"`
}

// ConceptInfo models concept metadata.
type ConceptInfo struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Difficulty  Difficulty `json:"difficulty"`
	Description string     `json:"description,omitempty"`
}
