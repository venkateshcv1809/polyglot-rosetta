package runcmd

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
	if command.Name != commandName || command.Summary != commandSummary {
		t.Fatalf("command = %#v", command)
	}
	if command.Run == nil {
		t.Fatal("Run is nil")
	}
}

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want options
	}{
		{
			name: "file input with defaults",
			args: []string{"--concept=hello-world", "--path=main.go"},
			want: options{concept: "hello-world", path: "main.go", language: defaultLanguage},
		},
		{
			name: "standard input with overrides",
			args: []string{"--concept=hello-world", "--stdin", "--lang=python", "--all"},
			want: options{concept: "hello-world", stdin: true, language: "python", all: true},
		},
		{
			name: "inline input",
			args: []string{"--concept=hello-world", "--code=package main"},
			want: options{concept: "hello-world", code: "package main", language: defaultLanguage},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseOptions("run", test.args)
			if err != nil {
				t.Fatalf("parseOptions: %v", err)
			}
			if got != test.want {
				t.Fatalf("options = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestParseOptionsValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "missing concept", args: []string{"--path=main.go"}, wantErr: "--concept"},
		{name: "missing input", args: []string{"--concept=hello-world"}, wantErr: "missing input source"},
		{name: "positional argument", args: []string{"source.go"}, wantErr: "unexpected argument"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseOptions("run", test.args)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, test.wantErr)
			}
		})
	}
}
