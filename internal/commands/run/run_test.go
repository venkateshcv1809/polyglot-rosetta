package runcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceCodeReadsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.go")
	want := "package main\n"
	if err := os.WriteFile(path, []byte(want), 0o600); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	got, err := sourceCode(options{path: path})
	if err != nil {
		t.Fatalf("sourceCode: %v", err)
	}
	if got != want {
		t.Fatalf("source = %q, want %q", got, want)
	}
}

func TestSourceCodeReportsMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.go")
	_, err := sourceCode(options{path: path})
	if err == nil || !strings.Contains(err.Error(), "failed to read source file") {
		t.Fatalf("error = %v, want source file read error", err)
	}
}

func TestRunDoesNotReportUnimplementedExecutionAsSuccess(t *testing.T) {
	err := run(options{concept: "hello-world", language: "go", code: "package main"})
	if err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("error = %v, want execution engine error", err)
	}
}
