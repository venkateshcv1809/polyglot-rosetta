package indexcmd

import (
	"errors"
	"flag"
	"testing"
)

func TestExecuteHelp(t *testing.T) {
	if err := Execute([]string{"--help"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("error = %v, want flag.ErrHelp", err)
	}
}

func TestCommand(t *testing.T) {
	command := Command()
	if command.Name != commandName || command.Summary != commandSummary {
		t.Fatalf("command = %#v", command)
	}
	if command.Run == nil {
		t.Fatal("Run is nil")
	}
}

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOutput string
	}{
		{name: "default output", wantOutput: defaultOutputPath()},
		{name: "custom output", args: []string{"--output=build/catalog.json"}, wantOutput: "build/catalog.json"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts, err := parseOptions("generate index", test.args)
			if err != nil {
				t.Fatalf("parseOptions: %v", err)
			}
			if opts.output != test.wantOutput {
				t.Fatalf("output = %q, want %q", opts.output, test.wantOutput)
			}
		})
	}
}
