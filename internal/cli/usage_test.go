package cli

import (
	"bytes"
	"testing"
)

func TestPrintUsage(t *testing.T) {
	var output bytes.Buffer
	PrintUsage(&output, "prosetta run", []string{
		"--id=<id> --lang=<language> --stdin [options]",
	}, []Option{
		{Name: "id", Value: "id", Description: "Concept identifier", Required: true},
		{Name: "lang", Value: "language", Description: "Programming language", Required: true},
	})

	want := "Usage:\n" +
		"  prosetta run --id=<id> --lang=<language> --stdin [options]\n" +
		"\nOptions:\n" +
		"  --id=<id>          Concept identifier (required)\n" +
		"  --lang=<language>  Programming language (required)\n"
	if output.String() != want {
		t.Fatalf("output:\n%q\nwant:\n%q", output.String(), want)
	}
}
