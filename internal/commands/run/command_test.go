package runcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"testing"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/domain/evaluation"
	"polyglot-rosetta/internal/toolchain"
)

func TestRunCommandReportsFailedBuildTiming(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("Go toolchain is unavailable")
	}
	cache := useRunTestCache(t).Toolchain
	if _, err := toolchain.RefreshCache(cache,
		func(name string) (string, error) { return "/tools/" + name, nil },
		func(string, string) (string, error) { return "fixture", nil },
	); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PROSETTA_CATALOG", writeCatalog(t))
	id := mustConceptID(t, "fundamentals/hello-world")
	for _, format := range []string{"json", "text"} {
		t.Run(format, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := runCommandWithIO(context.Background(), "prosetta run", []string{
				"--id=" + id, "--lang=go", "--code=invalid Go source", "--format=" + format,
			}, strings.NewReader(""), &stdout)
			if err == nil || cli.ExitCode(err, &stderr) != 1 {
				t.Fatalf("build failure did not fail command: %v", err)
			}
			if format == "json" {
				var result runapi.Result
				if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				if result.Status != evaluation.StatusError || result.BuildDurationMs == nil || *result.BuildDurationMs < 0 || result.DurationMs < *result.BuildDurationMs || len(result.TestResults) != 0 || result.Error == "" || stderr.Len() != 0 {
					t.Fatalf("failed build result = %+v, stderr = %q", result, stderr.String())
				}
			} else if !strings.Contains(stdout.String(), "Build Time:") || !strings.Contains(stderr.String(), "build command") {
				t.Fatalf("build timing or diagnostic lost: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
		})
	}
}

func TestCommand(t *testing.T) {
	command := Command()
	if command.Name != commandName || command.Summary != commandSummary {
		t.Fatalf("command = %#v", command)
	}
	if command.Run == nil {
		t.Fatal("Run is nil")
	}
}

func TestRunCommandJSONOptionErrors(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{name: "missing ID", args: []string{"--format=json", "--lang=go"}, want: "--id"},
		{name: "unknown flag before format", args: []string{"--unknown", "--format=json"}, want: "flag provided but not defined"},
		{name: "separate format value", args: []string{"--lang=go", "--format", "json"}, want: "--id"},
		{name: "invalid language", args: []string{"--id=c-2b23ca07", "--lang=ruby", "--format=json"}, want: "unsupported language"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := runCommandWithIO(context.Background(), "prosetta run", test.args, strings.NewReader(""), &stdout)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
			if code := cli.ExitCode(err, &stderr); code != 1 || stderr.Len() != 0 {
				t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
			}
			var result runapi.Result
			decoder := json.NewDecoder(&stdout)
			if err := decoder.Decode(&result); err != nil {
				t.Fatalf("decode JSON: %v; output = %q", err, stdout.String())
			}
			if err := decoder.Decode(new(any)); err != io.EOF {
				t.Fatalf("unexpected trailing output: %v", err)
			}
			if result.Status != evaluation.StatusError || !strings.Contains(result.Error, test.want) || result.TestResults == nil {
				t.Fatalf("result = %+v", result)
			}
		})
	}
}
