package execution

import (
	"fmt"
	"os"
	"path/filepath"
)

// Workspace manages dedicated temporary directories for execution sessions.
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

// WriteFile writes a raw source code string or asset into the workspace.
func (w *Workspace) WriteFile(filename, content string) (string, error) {
	if filename == "" || filepath.IsAbs(filename) || filepath.Base(filename) != filename {
		return "", fmt.Errorf("invalid workspace filename %q", filename)
	}
	filePath := filepath.Join(w.Dir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("failed to write file %s in workspace: %w", filename, err)
	}
	return filePath, nil
}
