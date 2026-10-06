package execution

import (
	"os"
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
