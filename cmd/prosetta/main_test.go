package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainHandlesHelp(t *testing.T) {
	location := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("PROSETTA_CONFIG", location)
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	os.Args = []string{"prosetta", "--help"}

	main()
	if _, err := os.Stat(location); !os.IsNotExist(err) {
		t.Fatalf("help created cache config: %v", err)
	}
}

func TestExecute_PublicSurface(t *testing.T) {
	if err := execute(nil); err == nil {
		t.Fatal("expected missing command")
	}
	for _, name := range []string{"generate", "site"} {
		if err := execute([]string{name}); err == nil {
			t.Fatalf("expected unknown command for %s", name)
		} else if !strings.Contains(err.Error(), "unknown command") {
			t.Fatalf("%s error = %v", name, err)
		}
	}
	if err := execute([]string{"--help"}); err != nil {
		t.Fatalf("help: %v", err)
	}
	if err := execute([]string{"catalog", "--help"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("catalog help: %v", err)
	}
	for _, name := range []string{"probe", "run", "version"} {
		if err := execute([]string{name, "--help"}); !errors.Is(err, flag.ErrHelp) {
			t.Errorf("%s help: %v", name, err)
		}
	}
}
