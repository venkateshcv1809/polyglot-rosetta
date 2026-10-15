package probecmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	probeapi "polyglot-rosetta/internal/api/probe"
	"polyglot-rosetta/internal/cli"
)

func TestProbeCommandJSONSuccess(t *testing.T) {
	var output bytes.Buffer
	err := probeCommandWithIO("prosetta probe", []string{"--lang=go", "--format=json"}, &output,
		func(string) (string, error) { return "/tools/go", nil },
		func(string, string) (string, error) { return "go 1.0", nil },
	)
	if err != nil {
		t.Fatalf("probeCommandWithIO: %v", err)
	}
	var result probeapi.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode JSON: %v; output = %q", err, output.String())
	}
	if result.Status != probeapi.StatusOK || result.Error != "" || len(result.Toolchains) != 1 ||
		result.Toolchains[0].Language != "go" || !result.Toolchains[0].Available ||
		len(result.Toolchains[0].Tools) != 1 || result.Toolchains[0].Tools[0].Version != "go 1.0" {
		t.Fatalf("result = %+v", result)
	}
}

func TestProbeCommandJSONUnavailable(t *testing.T) {
	var output, stderr bytes.Buffer
	err := probeCommandWithIO("prosetta probe", []string{"--lang=go", "--format", "json"}, &output,
		func(string) (string, error) { return "", errors.New("not found") },
		func(string, string) (string, error) { t.Fatal("unexpected inspect"); return "", nil },
	)
	if code := cli.ExitCode(err, &stderr); code != 1 || stderr.Len() != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	var result probeapi.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if result.Status != probeapi.StatusUnavailable || result.Error == "" || result.Toolchains[0].Available ||
		result.Toolchains[0].Tools[0].Error != "not found" {
		t.Fatalf("result = %+v", result)
	}
}

func TestProbeCommandJSONParseError(t *testing.T) {
	var output, stderr bytes.Buffer
	err := probeCommandWithIO("prosetta probe", []string{"--format=json", "--lang=ruby"}, &output,
		func(string) (string, error) { t.Fatal("unexpected lookup"); return "", nil },
		func(string, string) (string, error) { t.Fatal("unexpected inspect"); return "", nil },
	)
	if code := cli.ExitCode(err, &stderr); code != 1 || stderr.Len() != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	var result probeapi.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode JSON: %v; output = %q", err, output.String())
	}
	if result.Status != probeapi.StatusError || !strings.Contains(result.Error, "unsupported language") || result.Toolchains == nil || len(result.Toolchains) != 0 {
		t.Fatalf("result = %+v", result)
	}
}
