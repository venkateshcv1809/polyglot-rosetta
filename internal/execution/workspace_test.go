package execution

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceCreationAndCleanup(t *testing.T) {
	ws, err := NewWorkspace()
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	if ws.Dir == "" {
		t.Errorf("expected workspace directory path to be non-empty")
	}

	// Verify the directory actually exists on disk
	info, err := os.Stat(ws.Dir)
	if err != nil || !info.IsDir() {
		t.Errorf("workspace directory does not exist or is not a directory: %v", err)
	}

	// Test writing a source string into the workspace
	filePath, err := ws.WriteFile("solution.go", "package main\n")
	if err != nil {
		t.Errorf("failed to write source file: %v", err)
	}

	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("written source file does not exist: %v", err)
	}

	// Test cleanup
	if err := ws.Cleanup(); err != nil {
		t.Errorf("failed to cleanup workspace: %v", err)
	}

	if _, err := os.Stat(ws.Dir); !os.IsNotExist(err) {
		t.Errorf("workspace directory was not removed after cleanup")
	}
}

func TestWorkspaceRejectsPathsOutsideRoot(t *testing.T) {
	ws, err := NewWorkspace()
	if err != nil {
		t.Fatalf("NewWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = ws.Cleanup() })

	for _, name := range []string{"", "../outside.go", filepath.Join("nested", "file.go"), "/tmp/file.go"} {
		if _, err := ws.WriteFile(name, "content"); err == nil {
			t.Errorf("WriteFile(%q) returned nil error", name)
		}
	}
}
