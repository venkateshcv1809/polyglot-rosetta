package probecmd

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

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
		{name: "all languages explicitly", args: []string{"--lang=all"}},
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

func TestParseOptionsRejectsUnsupportedLanguage(t *testing.T) {
	_, err := parseOptions("probe", []string{"--lang=ruby"})
	if err == nil || !strings.Contains(err.Error(), "unsupported language") {
		t.Fatalf("error = %v, want unsupported language", err)
	}
}

func TestProbeReportsAllRequiredBinaries(t *testing.T) {
	useTestCache(t)
	var output bytes.Buffer
	lookPath := func(binary string) (string, error) {
		if binary == "node" {
			return "", fmt.Errorf("not found")
		}
		return "/toolchains/" + binary, nil
	}
	inspect := func(binary, path string) (string, error) { return binary + " 1.0", nil }
	if err := probeWithInspector(options{}, &output, lookPath, inspect); !errors.Is(err, errToolchainsUnavailable) {
		t.Fatalf("probeWithInspector error = %v, want unavailable", err)
	}
	got := output.String()
	if !strings.Contains(got, "OK    go") || !strings.Contains(got, "FAIL  typescript") || !strings.Contains(got, `missing "node"`) {
		t.Fatalf("output:\n%s", got)
	}
}

func TestProbeSelectedLanguageOutput(t *testing.T) {
	useTestCache(t)
	var output bytes.Buffer
	err := probeWithInspector(
		options{language: "typescript"},
		&output,
		func(binary string) (string, error) { return "/tools/" + binary, nil },
		func(binary, path string) (string, error) { return binary + " 1.0", nil },
	)
	if err != nil {
		t.Fatalf("probeWithInspector: %v", err)
	}
	want := "\nToolchains:\n" +
		"  OK    typescript    /tools/tsc [tsc 1.0], /tools/node [node 1.0]\n"
	if got := output.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestProbePropagatesOutputFailure(t *testing.T) {
	useTestCache(t)
	want := errors.New("write failed")
	err := probeWithInspector(
		options{language: "go"},
		failingWriter{err: want},
		func(binary string) (string, error) { return "/tools/" + binary, nil },
		func(binary, path string) (string, error) { return "go 1.0", nil },
	)
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want write failure", err)
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestProbeRejectsBinaryThatCannotReportVersion(t *testing.T) {
	useTestCache(t)
	var output bytes.Buffer
	err := probeWithInspector(
		options{language: "go"},
		&output,
		func(binary string) (string, error) { return "/toolchains/" + binary, nil },
		func(binary, path string) (string, error) { return "", errors.New("broken") },
	)
	if err == nil || !strings.Contains(output.String(), "unusable") {
		t.Fatalf("error = %v, output = %q", err, output.String())
	}
}

func useTestCache(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", base)
	t.Setenv("PROSETTA_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	return filepath.Join(base, "prosetta", "toolchains.json")
}
