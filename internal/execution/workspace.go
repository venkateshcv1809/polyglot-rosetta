package execution

import (
	"fmt"
	"os"
	"path/filepath"
)

// Workspace manages isolated temporary directories for individual execution sessions.
type Workspace struct {
	Dir string
}

// NewWorkspace creates a unique temporary directory using OS temp APIs.
func NewWorkspace() (*Workspace, error) {
	dir, err := os.MkdirTemp("", "polyglot-rosetta-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary workspace: %w", err)
	}
	return &Workspace{Dir: dir}, nil
}

// Cleanup removes the temporary workspace directory and all its contents.
func (w *Workspace) Cleanup() error {
	if w.Dir == "" {
		return nil
	}
	return os.RemoveAll(w.Dir)
}

// WriteFile writes a raw source code string or asset into the isolated workspace.
func (w *Workspace) WriteFile(filename, content string) (string, error) {
	filePath := filepath.Join(w.Dir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s in workspace: %w", filename, err)
	}
	return filePath, nil
}
