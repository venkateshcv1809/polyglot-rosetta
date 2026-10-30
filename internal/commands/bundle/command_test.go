package bundlecmd

import (
	"errors"
	"flag"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommand(t *testing.T) {
	command := Command()
	if command.Name != commandName || command.Summary != commandSummary || command.Run == nil {
		t.Fatalf("command = %#v", command)
	}
}

func TestParseOptionsRejectsEmptyOutput(t *testing.T) {
	for _, value := range []string{"", "   "} {
		if _, err := parseOptions("generate bundle", []string{"--output=" + value}); err == nil || !strings.Contains(err.Error(), "cannot be empty") {
			t.Fatalf("output %q: error = %v", value, err)
		}
	}
}

func TestParseOptions(t *testing.T) {
	opts, err := parseOptions("generate bundle", []string{"--output=custom"})
	if err != nil || opts.output != "custom" {
		t.Fatalf("options = %+v, error = %v", opts, err)
	}
	opts, err = parseOptions("generate bundle", nil)
	if err != nil || opts.output != filepath.Join("dist", "problems") {
		t.Fatalf("default options = %+v, error = %v", opts, err)
	}
	if _, err := parseOptions("generate bundle", []string{"-h"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("help error = %v", err)
	}
}
