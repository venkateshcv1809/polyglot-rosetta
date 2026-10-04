package probecmd

import (
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestExecuteHelp(t *testing.T) {
	if err := Execute([]string{"--help"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("error = %v, want flag.ErrHelp", err)
	}
}

func TestCommand(t *testing.T) {
	command := Command()
	if command.Name != commandName {
		t.Fatalf("name = %q, want %q", command.Name, commandName)
	}
	if command.Summary != commandSummary {
		t.Fatalf("summary = %q, want %q", command.Summary, commandSummary)
	}
	if command.Run == nil {
		t.Fatal("Run is nil")
	}
	if len(command.Subcommands) != 0 {
		t.Fatalf("subcommands = %d, want 0", len(command.Subcommands))
	}
}

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantLanguage string
	}{
		{name: "all languages by default"},
		{name: "selected language", args: []string{"--lang=python"}, wantLanguage: "python"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts, err := parseOptions("probe", test.args)
			if err != nil {
				t.Fatalf("parseOptions: %v", err)
			}
			if opts.language != test.wantLanguage {
				t.Fatalf("language = %q, want %q", opts.language, test.wantLanguage)
			}
		})
	}
}

func TestParseOptionsRejectsPositionalArgument(t *testing.T) {
	_, err := parseOptions("probe", []string{"python"})
	if err == nil || !strings.Contains(err.Error(), "unexpected argument") {
		t.Fatalf("error = %v, want unexpected argument", err)
	}
}

func TestBinaryName(t *testing.T) {
	tests := map[string]string{
		"go":         "go",
		"python":     "python3",
		"rust":       "rustc",
		"typescript": "tsc",
		"ts":         "tsc",
		"zig":        "zig",
	}

	for language, want := range tests {
		if got := binaryName(language); got != want {
			t.Errorf("binaryName(%q) = %q, want %q", language, got, want)
		}
	}
}
