package main

import (
	"errors"
	"flag"
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
	if err := execute([]string{"catalog", "--help"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("catalog help: %v", err)
	}
	for _, name := range []string{"probe", "run"} {
		if err := execute([]string{name, "--help"}); !errors.Is(err, flag.ErrHelp) {
			t.Errorf("%s help: %v", name, err)
		}
	}
}
