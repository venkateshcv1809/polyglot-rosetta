package probecmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	probeapi "polyglot-rosetta/internal/api/probe"
	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/toolchain"
)

func TestProbeCommandJSONSuccess(t *testing.T) {
	cachePath := useTestCache(t)
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
	if err := toolchain.RequireAvailable(cachePath, language.Go, func(string) (string, error) {
		t.Fatal("manual probe was not cached")
		return "", nil
	}, func(string, string) (string, error) {
		t.Fatal("manual probe was not cached")
		return "", nil
	}); err != nil {
		t.Fatalf("read probe cache: %v", err)
	}
	if err := toolchain.RequireAvailable(cachePath, language.TypeScript, func(string) (string, error) {
		t.Fatal("selected-language probe did not cache all languages")
		return "", nil
	}, func(string, string) (string, error) {
		t.Fatal("selected-language probe did not cache all languages")
		return "", nil
	}); err != nil {
		t.Fatalf("read TypeScript cache: %v", err)
	}
}

func TestProbeCommandJSONUnavailable(t *testing.T) {
	cachePath := useTestCache(t)
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
	if err := toolchain.RequireAvailable(cachePath, language.Go, func(string) (string, error) {
		t.Fatal("cached unavailable result performed lookup")
		return "", nil
	}, func(string, string) (string, error) {
		t.Fatal("cached unavailable result inspected executable")
		return "", nil
	}); !errors.Is(err, toolchain.ErrUnavailable) {
		t.Fatalf("cached availability error = %v", err)
	}
}

func TestProbeCommandJSONParseError(t *testing.T) {
	useTestCache(t)
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

func TestProbeCommandReportsCacheWriteFailureAsJSON(t *testing.T) {
	cachePath := useTestCache(t)
	parent := filepath.Dir(cachePath)
	if err := os.WriteFile(parent, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	var output, stderr bytes.Buffer
	err := probeCommandWithIO("prosetta probe", []string{"--lang=go", "--format=json"}, &output,
		func(binary string) (string, error) { return "/tools/" + binary, nil },
		func(string, string) (string, error) { return "go 1.0", nil },
	)
	if code := cli.ExitCode(err, &stderr); code != 1 || stderr.Len() != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	var result probeapi.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if result.Status != probeapi.StatusError || !strings.Contains(result.Error, "save toolchain cache") {
		t.Fatalf("result = %+v", result)
	}
}

func TestProbeClearDoesNotInspectAndNextRunCanRebuild(t *testing.T) {
	cachePath := useTestCache(t)
	if _, err := toolchain.RefreshCache(cachePath,
		func(binary string) (string, error) { return "/tools/" + binary, nil },
		func(string, string) (string, error) { return "1.0", nil },
	); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := probeCommandWithIO("prosetta probe", []string{"--clear", "--format=json"}, &output,
		func(string) (string, error) { t.Fatal("clear performed lookup"); return "", nil },
		func(string, string) (string, error) { t.Fatal("clear inspected executable"); return "", nil },
	); err != nil {
		t.Fatal(err)
	}
	var result probeapi.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != probeapi.StatusCleared || result.Toolchains == nil || len(result.Toolchains) != 0 {
		t.Fatalf("result = %+v", result)
	}
	if _, err := os.Stat(cachePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stat cache: %v", err)
	}
	var textOutput bytes.Buffer
	if err := probeCommandWithIO("prosetta probe", []string{"--clear"}, &textOutput,
		func(string) (string, error) { t.Fatal("clear performed lookup"); return "", nil },
		func(string, string) (string, error) { t.Fatal("clear inspected executable"); return "", nil },
	); err != nil || textOutput.String() != "\nOK:\n  Toolchain cache cleared\n" {
		t.Fatalf("clear text = %q, error = %v", textOutput.String(), err)
	}
}

func TestProbeClearReportsFailureAsJSON(t *testing.T) {
	cachePath := useTestCache(t)
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		t.Fatal(err)
	}
	var output, stderr bytes.Buffer
	err := probeCommandWithIO("prosetta probe", []string{"--clear", "--format=json"}, &output,
		func(string) (string, error) { t.Fatal("clear performed lookup"); return "", nil },
		func(string, string) (string, error) { t.Fatal("clear inspected executable"); return "", nil },
	)
	if code := cli.ExitCode(err, &stderr); code != 1 || stderr.Len() != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	var result probeapi.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != probeapi.StatusError || !strings.Contains(result.Error, "not a file") {
		t.Fatalf("result = %+v", result)
	}
}

func TestProbeClearTextFailureHasNoToolchainHeading(t *testing.T) {
	cachePath := useTestCache(t)
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		t.Fatal(err)
	}
	var output, stderr bytes.Buffer
	err := probeCommandWithIO("prosetta probe", []string{"--clear"}, &output,
		func(string) (string, error) { t.Fatal("clear performed lookup"); return "", nil },
		func(string, string) (string, error) { t.Fatal("clear inspected executable"); return "", nil },
	)
	if code := cli.ExitCode(err, &stderr); code != 1 || output.Len() != 0 || !strings.Contains(stderr.String(), "Toolchain cache not cleared") {
		t.Fatalf("exit code = %d, stdout = %q, stderr = %q", code, output.String(), stderr.String())
	}
}
