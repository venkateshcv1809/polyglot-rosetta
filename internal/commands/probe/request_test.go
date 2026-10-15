package probecmd

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestParseOptionsFormats(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{name: "default", want: "text"},
		{name: "text", args: []string{"--format=text"}, want: "text"},
		{name: "json", args: []string{"--format", "json"}, want: "json"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			opts, err := parseOptionsWithOutput("prosetta probe", test.args, &output)
			if err != nil || opts.format != test.want || output.Len() != 0 {
				t.Fatalf("options = %+v, output = %q, error = %v", opts, output.String(), err)
			}
		})
	}
}

func TestParseOptionsRejectsFormat(t *testing.T) {
	var output bytes.Buffer
	_, err := parseOptionsWithOutput("prosetta probe", []string{"--format=yaml"}, &output)
	if err == nil || !strings.Contains(err.Error(), "unsupported output format") || !strings.Contains(output.String(), "prosetta probe") {
		t.Fatalf("error = %v, output = %q", err, output.String())
	}
}

func TestParseOptionsSuppressesUsageForJSONErrors(t *testing.T) {
	for _, args := range [][]string{
		{"--lang=ruby", "--format=json"},
		{"--format", "json", "--unknown"},
	} {
		var output bytes.Buffer
		if _, err := parseOptionsWithOutput("prosetta probe", args, &output); err == nil || output.Len() != 0 {
			t.Fatalf("args = %q, error = %v, output = %q", args, err, output.String())
		}
	}
}

func TestParseOptionsHelpStaysText(t *testing.T) {
	var output bytes.Buffer
	_, err := parseOptionsWithOutput("prosetta probe", []string{"--format=json", "--help"}, &output)
	if !errors.Is(err, flag.ErrHelp) || !strings.Contains(output.String(), "Usage:") {
		t.Fatalf("error = %v, output = %q", err, output.String())
	}
}

func TestParseOptionsClear(t *testing.T) {
	var output bytes.Buffer
	opts, err := parseOptionsWithOutput("prosetta probe", []string{"--clear", "--format=json"}, &output)
	if err != nil || !opts.clear || opts.format != "json" || output.Len() != 0 {
		t.Fatalf("options = %+v, output = %q, error = %v", opts, output.String(), err)
	}
	_, err = parseOptionsWithOutput("prosetta probe", []string{"--clear", "--lang=go", "--format=json"}, &output)
	if err == nil || !strings.Contains(err.Error(), "cannot be combined") || output.Len() != 0 {
		t.Fatalf("error = %v, output = %q", err, output.String())
	}
}
