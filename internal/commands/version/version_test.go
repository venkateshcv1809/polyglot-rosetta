package versioncmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"strings"
	"testing"

	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/version"
)

func TestVersionCommand(t *testing.T) {
	var output bytes.Buffer
	if err := versionCommandWithIO("prosetta version", nil, &output); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(output.String()); got != version.Value {
		t.Fatalf("version = %q, want %q", got, version.Value)
	}
}

func TestVersionCommandRejectsArguments(t *testing.T) {
	if err := versionCommandWithIO("prosetta version", []string{"extra"}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected unexpected argument")
	}
}

func TestVersionCommandFormats(t *testing.T) {
	for _, args := range [][]string{{"--format=text"}, {"--format=json"}, {"--format", "json"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var output bytes.Buffer
			if err := versionCommandWithIO("prosetta version", args, &output); err != nil {
				t.Fatal(err)
			}
			if args[0] == "--format=text" {
				if output.String() != version.Value+"\n" {
					t.Fatalf("text output = %q", output.String())
				}
				return
			}
			var result map[string]any
			if err := json.Unmarshal(output.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if len(result) != 3 || result["schema_version"] != float64(1) || result["status"] != "ok" || result["version"] != version.Value {
				t.Fatalf("JSON result = %#v", result)
			}
		})
	}
}

func TestVersionCommandJSONErrors(t *testing.T) {
	for _, extra := range []string{"extra", "--unknown"} {
		t.Run(extra, func(t *testing.T) {
			var output, stderr bytes.Buffer
			err := versionCommandWithIO("prosetta version", []string{"--format=json", extra}, &output)
			if err == nil || cli.ExitCode(err, &stderr) != 1 || stderr.Len() != 0 {
				t.Fatalf("error = %v, stderr = %q", err, stderr.String())
			}
			var result map[string]any
			if err := json.Unmarshal(output.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if len(result) != 3 || result["schema_version"] != float64(1) || result["status"] != "error" || result["error"] == "" || result["error"] == nil {
				t.Fatalf("JSON error result = %#v", result)
			}
		})
	}
}

func TestVersionCommandInvalidFormat(t *testing.T) {
	err := versionCommandWithIO("prosetta version", []string{"--format=yaml"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "unsupported output format") {
		t.Fatalf("error = %v", err)
	}
}

func TestVersionCommandHelp(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"--format=json", "--help"}, {"--help", "--format=json"}} {
		var output bytes.Buffer
		err := versionCommandWithIO("prosetta version", args, &output)
		if !errors.Is(err, flag.ErrHelp) || !strings.Contains(output.String(), "--format") {
			t.Fatalf("args = %v: error = %v, output = %q", args, err, output.String())
		}
	}
}

func TestVersionCommandOutputFailure(t *testing.T) {
	for _, args := range [][]string{nil, {"--format=json"}} {
		if err := versionCommandWithIO("prosetta version", args, failingWriter{}); !errors.Is(err, io.ErrClosedPipe) {
			t.Fatalf("args = %v: error = %v", args, err)
		}
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }
