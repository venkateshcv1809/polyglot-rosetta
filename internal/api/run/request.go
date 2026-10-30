// Package run defines the transport-neutral contract for evaluating a solution.
package run

import (
	"fmt"
	"strings"

	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/domain/language"
)

// Mode controls which testcases participate in an evaluation.
type Mode string

const (
	// ModePublic evaluates only cases intended for normal practice.
	ModePublic Mode = "public"
	// ModeAll evaluates every locally available testcase.
	ModeAll Mode = "all"
)

// Request is the transport-neutral input accepted by native and browser hosts.
type Request struct {
	ID       string `json:"id"`
	Language string `json:"language"`
	Source   string `json:"source"`
	Mode     Mode   `json:"mode"`
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
	if r.Mode != ModePublic && r.Mode != ModeAll {
		return fmt.Errorf("mode must be %q or %q", ModePublic, ModeAll)
	}
	return nil
}
