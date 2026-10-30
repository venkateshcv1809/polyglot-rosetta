package cli

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestFlagParser(t *testing.T) {
	var output bytes.Buffer
	var language string
	var all bool
	parser := NewFlagParser("prosetta run", "[options]")
	parser.SetOutput(&output)
	parser.StringVar(&language, "go", Option{Name: "lang", Value: "language", Description: "Target language"})
	parser.BoolVar(&all, false, Option{Name: "all", Description: "Include hidden tests"})

	if err := parser.Parse([]string{"--lang=python", "--all"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if language != "python" || !all {
		t.Fatalf("language = %q, all = %v", language, all)
	}
	parser.PrintUsage()
	if !strings.Contains(output.String(), "Target language (default: go)") {
		t.Fatalf("default not derived from registration:\n%s", output.String())
	}
}

func TestFlagParserHelpAliases(t *testing.T) {
	for _, help := range []string{"help", "-h", "--help"} {
		t.Run(help, func(t *testing.T) {
			var output bytes.Buffer
			parser := NewFlagParser("probe", "[options]")
			parser.SetOutput(&output)
			var language string
			parser.StringVar(&language, "", Option{Name: "lang", Value: "language", Description: "Language to probe"})

			if err := parser.Parse([]string{help}); !errors.Is(err, flag.ErrHelp) {
				t.Fatalf("Parse error = %v, want flag.ErrHelp", err)
			}
			want := "Usage:\n  probe [options]\n\nOptions:\n  --lang=<language>  Language to probe\n"
			if output.String() != want {
				t.Fatalf("help output:\n%q\nwant:\n%q", output.String(), want)
			}
		})
	}
}

func TestFlagParserRejectsPositionalArguments(t *testing.T) {
	var output bytes.Buffer
	parser := NewFlagParser("probe", "[options]")
	parser.SetOutput(&output)

	if err := parser.Parse([]string{"python"}); err == nil || !strings.Contains(err.Error(), "unexpected argument") {
		t.Fatalf("Parse error = %v", err)
	}
}
