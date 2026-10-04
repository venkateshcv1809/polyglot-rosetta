package cli

import (
	"bytes"
	"testing"
)

func TestPrintUsage(t *testing.T) {
	var output bytes.Buffer
	PrintUsage(&output, "prosetta run", []string{
		"--concept=<id> --path=<file> [options]",
	}, []Option{
		{Name: "concept", Value: "id", Description: "Concept identifier", Required: true},
		{Name: "lang", Value: "language", Description: "Programming language", Default: "go"},
	})

	want := "Usage:\n" +
		"  prosetta run --concept=<id> --path=<file> [options]\n" +
		"\nOptions:\n" +
		"  --concept=<id>     Concept identifier (required)\n" +
		"  --lang=<language>  Programming language (default: go)\n"
	if output.String() != want {
		t.Fatalf("output:\n%q\nwant:\n%q", output.String(), want)
	}
}
