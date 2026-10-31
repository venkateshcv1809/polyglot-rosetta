// Package run defines the transport-neutral contract for evaluating a solution.
package run

import (
	"fmt"
	"strings"

	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/domain/evaluation"
	"polyglot-rosetta/internal/domain/language"
)

// Request is the transport-neutral input accepted by native and browser hosts.
type Request struct {
	ID       string          `json:"id"`
	Language string          `json:"language"`
	Source   string          `json:"source"`
	Mode     evaluation.Mode `json:"mode"`
}

// Validate reports whether the request can be passed to an execution backend.
func (r Request) Validate() error {
	if _, err := catalog.NormalizeConceptID(r.ID); err != nil {
		return err
	}
	if _, err := language.Parse(r.Language); err != nil {
		return fmt.Errorf("unsupported language %q", r.Language)
	}
	if strings.TrimSpace(r.Source) == "" {
		return fmt.Errorf("source code cannot be empty")
	}
	if r.Mode != evaluation.ModePublic && r.Mode != evaluation.ModeAll {
		return fmt.Errorf("mode must be %q or %q", evaluation.ModePublic, evaluation.ModeAll)
	}
	return nil
}
