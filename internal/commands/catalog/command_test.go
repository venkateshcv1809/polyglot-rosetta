package catalogcmd

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestCommand(t *testing.T) {
	command := Command()
	if command.Name != commandName || command.Summary != commandSummary || command.Run == nil || len(command.Subcommands) != 0 {
		t.Fatalf("command = %#v", command)
	}
}

func TestParseOptions(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG_URL", "")
	if _, err := parseOptions("prosetta catalog", nil); err == nil || !strings.Contains(err.Error(), "--url") {
		t.Fatalf("missing URL error = %v", err)
	}
	if _, err := parseOptions("prosetta catalog", []string{"-h"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("help error = %v", err)
	}
	opts, err := parseOptions("prosetta catalog", []string{"--url=https://example.test/problems/", "--cache=/tmp/catalog"})
	if err != nil || opts.url != "https://example.test/problems/" || opts.cache != "/tmp/catalog" {
		t.Fatalf("options = %+v, error = %v", opts, err)
	}
}

func TestParseOptionsFormatsAndEnvironment(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG_URL", "https://static.example/problems/")
	var output bytes.Buffer
	opts, err := parseOptionsWithOutput("prosetta catalog", []string{"--format=json"}, &output)
	if err != nil || opts.url != "https://static.example/problems/" || opts.format != "json" || output.Len() != 0 {
		t.Fatalf("options = %+v, output = %q, error = %v", opts, output.String(), err)
	}
	if _, err := parseOptionsWithOutput("prosetta catalog", []string{"--format=yaml"}, &output); err == nil || !strings.Contains(err.Error(), "unsupported output format") {
		t.Fatalf("format error = %v", err)
	}
	if _, err := parseOptionsWithOutput("prosetta catalog", []string{"--cache=", "--format=json"}, &output); err == nil || !strings.Contains(err.Error(), "cache directory cannot be empty") {
		t.Fatalf("cache error = %v", err)
	}
}

func TestParseOptionsJSONErrorsDoNotPrintUsage(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG_URL", "")
	for _, args := range [][]string{
		{"--format=json"},
		{"--format", "json", "--unknown"},
	} {
		var output bytes.Buffer
		if _, err := parseOptionsWithOutput("prosetta catalog", args, &output); err == nil || output.Len() != 0 {
			t.Fatalf("args = %q, output = %q, error = %v", args, output.String(), err)
		}
	}
	var output bytes.Buffer
	if _, err := parseOptionsWithOutput("prosetta catalog", []string{"--format=json", "--help"}, &output); !errors.Is(err, flag.ErrHelp) || !strings.Contains(output.String(), "Usage:") {
		t.Fatalf("help output = %q, error = %v", output.String(), err)
	}
}
