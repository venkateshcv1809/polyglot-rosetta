package execution

import (
	"os"
	"path/filepath"
	"testing"

	"polyglot-rosetta/internal/domain/language"
)

func TestResolveTemplatePath_LocalOverride(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "concept_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	templatesDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("failed to create templates dir: %v", err)
	}

	localFile := filepath.Join(templatesDir, "solution.go")
	if err := os.WriteFile(localFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to write mock template: %v", err)
	}

	resolved, err := ResolveTemplatePath(tmpDir, language.Go, "/nonexistent/global")
	if err != nil {
		t.Fatalf("expected successful resolution, got error: %v", err)
	}

	if resolved != localFile {
		t.Errorf("expected local override path %s, got %s", localFile, resolved)
	}
}
