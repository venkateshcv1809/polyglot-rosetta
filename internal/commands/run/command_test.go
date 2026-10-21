package runcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/domain/evaluation"
)

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
