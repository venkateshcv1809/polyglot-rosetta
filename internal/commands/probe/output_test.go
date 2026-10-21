package probecmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	probeapi "polyglot-rosetta/internal/api/probe"
)

func TestWriteTextResultFromTemplate(t *testing.T) {
	result := probeapi.Result{Toolchains: []probeapi.Toolchain{
		{Language: "go", Available: true, Tools: []probeapi.Tool{{Binary: "go", Available: true, Path: "/tools/go", Version: "go 1.0"}}},
		{Language: "typescript", Tools: []probeapi.Tool{
			{Binary: "tsc", Path: "/tools/tsc", Error: "unusable"},
			{Binary: "node", Error: "not found"},
		}},
	}}
	var output bytes.Buffer
	if err := writeResult(&output, "text", result); err != nil {
		t.Fatalf("writeResult: %v", err)
	}
	want := "\nToolchains:\n" +
		"  OK    go            /tools/go [go 1.0]\n" +
		"  FAIL  typescript    missing \"tsc (unusable)\", \"node\"\n"
	if got := output.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestWriteJSONResult(t *testing.T) {
	result := probeapi.Result{Status: probeapi.StatusOK, Toolchains: []probeapi.Toolchain{}}
	var output bytes.Buffer
	if err := writeResult(&output, "json", result); err != nil {
		t.Fatalf("writeResult: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if decoded["status"] != "ok" || decoded["toolchains"] == nil || !strings.HasSuffix(output.String(), "\n") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestWriteJSONResultPropagatesWriteError(t *testing.T) {
	want := errors.New("write failed")
	err := writeResult(failingWriter{err: want}, "json", probeapi.Result{Status: probeapi.StatusOK})
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want write failure", err)
	}
}
