package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainHandlesHelp(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	os.Args = []string{"prosetta", "--help"}

	main()
}

func TestExecute_PublicSurface(t *testing.T) {
	if err := execute(nil); err == nil {
		t.Fatal("expected missing command")
	}
	if err := execute([]string{"generate"}); err == nil {
		t.Fatal("expected unknown command for generate")
	} else if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("error = %v", err)
	}
	if err := execute([]string{"--help"}); err != nil {
		t.Fatalf("help: %v", err)
	}
}
