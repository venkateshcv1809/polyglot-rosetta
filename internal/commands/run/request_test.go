package runcmd

import (
	"strings"
	"testing"

	"polyglot-rosetta/internal/catalog"
)

func TestParseOptions(t *testing.T) {
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	tests := []struct {
		name string
		args []string
		want options
	}{
		{
			name: "catalog solution",
			args: []string{"--id=" + id, "--lang=go"},
			want: options{id: id, language: "go", format: "text"},
		},
		{
			name: "standard input",
			args: []string{"--id=" + id, "--stdin", "--lang=go"},
			want: options{id: id, stdin: true, language: "go", format: "text"},
		},
		{
			name: "standard input with overrides",
			args: []string{"--id=" + id[:10], "--stdin", "--lang=python", "--all"},
			want: options{id: id[:10], stdin: true, language: "python", format: "text", all: true},
		},
		{
			name: "inline input",
			args: []string{"--id=" + id, "--code=package main", "--lang=go"},
			want: options{id: id, code: "package main", language: "go", format: "text"},
		},
		{
			name: "JSON output",
			args: []string{"--id=" + id, "--lang=go", "--format=json"},
			want: options{id: id, language: "go", format: "json"},
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
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "missing ID", args: []string{"--stdin", "--lang=go"}, wantErr: "--id"},
		{name: "missing language", args: []string{"--id=" + id, "--stdin"}, wantErr: "--lang"},
		{name: "invalid ID", args: []string{"--id=hello-world", "--lang=go", "--stdin"}, wantErr: "concept ID"},
		{name: "multiple inputs", args: []string{"--id=" + id, "--lang=go", "--code=source", "--stdin"}, wantErr: "multiple input sources"},
		{name: "unsupported language", args: []string{"--id=" + id, "--code=source", "--lang=ruby"}, wantErr: "unsupported language"},
		{name: "unsupported format", args: []string{"--id=" + id, "--lang=go", "--format=yaml"}, wantErr: "unsupported output format"},
		{name: "old path option", args: []string{"--id=" + id, "--lang=go", "--path=main.go"}, wantErr: "flag provided but not defined"},
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
