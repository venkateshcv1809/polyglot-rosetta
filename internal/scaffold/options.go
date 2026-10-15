package scaffold

import (
	"errors"
	"fmt"
	"regexp"

	"polyglot-rosetta/internal/domain/catalog"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// CreateCategoryOptions contains the metadata and parent path for a category.
type CreateCategoryOptions struct {
	ParentPath  string
	ID          string
	Title       string
	Description string
}

// Validate verifies that category options can produce a valid catalog node.
func (o CreateCategoryOptions) Validate() error {
	if o.ID == "" {
		return errors.New("category ID cannot be empty")
	}
	if !slugRegex.MatchString(o.ID) {
		return fmt.Errorf("invalid category ID %q: must be lowercase kebab-case (e.g. 'control-flow')", o.ID)
	}
	if o.Title == "" {
		return errors.New("category title cannot be empty")
	}
	return nil
}

// CreateConceptOptions contains the metadata and parent path for a concept.
type CreateConceptOptions struct {
	ParentPath  string
	ID          string
	Title       string
	Difficulty  catalog.Difficulty
	Description string
}

// Validate verifies that concept options can produce a valid catalog node.
func (o CreateConceptOptions) Validate() error {
	if o.ID == "" {
		return errors.New("concept ID cannot be empty")
	}
	if !slugRegex.MatchString(o.ID) {
		return fmt.Errorf("invalid concept ID %q: must be lowercase kebab-case (e.g. 'hello-world')", o.ID)
	}
	if o.Title == "" {
		return errors.New("concept title cannot be empty")
	}
	if err := o.Difficulty.Validate(); err != nil {
		return err
	}
	return nil
}
