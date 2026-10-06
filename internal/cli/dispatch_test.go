package cli

import (
	"bytes"
	"errors"
	"flag"
	"testing"
)

func TestExecuteRoutesNestedCommandWithFullPath(t *testing.T) {
	var gotPath string
	var gotArgs []string
	root := Command{
		Name:    "prosetta",
		Summary: "public CLI",
		Subcommands: []Command{{
			Name:    "generate",
			Summary: "generate content",
			Subcommands: []Command{{
				Name:    "category",
				Summary: "scaffold category",
				Run: func(commandPath string, args []string) error {
					gotPath = commandPath
					gotArgs = append([]string(nil), args...)
					return nil
				},
			}},
		}},
	}

	if err := Execute(root, []string{"generate", "category", "--id=sorting"}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gotPath != "prosetta generate category" {
		t.Fatalf("command path = %q", gotPath)
	}
	if len(gotArgs) != 1 || gotArgs[0] != "--id=sorting" {
		t.Fatalf("args = %v", gotArgs)
	}
}

func TestExecuteLeaf(t *testing.T) {
	want := errors.New("sentinel")
	command := Command{Name: "run", Run: func(path string, args []string) error {
		if path != "run" {
			t.Fatalf("path = %q", path)
		}
		return want
	}}
	if err := Execute(command, nil); !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
}

func TestExecuteMissingUnknownAndHelp(t *testing.T) {
	root := Command{Name: "prosetta", Subcommands: []Command{{Name: "run", Run: func(string, []string) error { return nil }}}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := ExecuteWithWriters(root, nil, &stdout, &stderr); err == nil {
		t.Fatal("expected missing command error")
	}
	wantUsage := "Usage:\n  prosetta <command> [options]\n\nCommands:\n  run  \n"
	if stderr.String() != wantUsage {
		t.Fatalf("missing-command output:\n%q\nwant:\n%q", stderr.String(), wantUsage)
	}

	stderr.Reset()
	if err := ExecuteWithWriters(root, []string{"unknown"}, &stdout, &stderr); err == nil {
		t.Fatal("expected unknown command error")
	}
	if stderr.String() != wantUsage {
		t.Fatalf("unknown-command output:\n%q\nwant:\n%q", stderr.String(), wantUsage)
	}

	for _, help := range []string{"help", "-h", "--help"} {
		stdout.Reset()
		if err := ExecuteWithWriters(root, []string{help}, &stdout, &stderr); err != nil {
			t.Fatalf("%s: %v", help, err)
		}
		if stdout.String() != wantUsage {
			t.Fatalf("%s output:\n%q\nwant:\n%q", help, stdout.String(), wantUsage)
		}
	}
}

func TestValidateCommand(t *testing.T) {
	validLeaf := func(name string) Command {
		return Command{Name: name, Run: func(string, []string) error { return nil }}
	}
	tests := []struct {
		name    string
		command Command
		wantErr string
	}{
		{name: "empty name", command: Command{Run: func(string, []string) error { return nil }}, wantErr: "name cannot be empty"},
		{name: "whitespace in name", command: validLeaf("bad name"), wantErr: "cannot contain whitespace"},
		{name: "missing handler", command: Command{Name: "run"}, wantErr: "has no handler"},
		{
			name:    "handler and children",
			command: Command{Name: "root", Run: func(string, []string) error { return nil }, Subcommands: []Command{validLeaf("run")}},
			wantErr: "both subcommands and a handler",
		},
		{
			name:    "duplicate child",
			command: Command{Name: "root", Subcommands: []Command{validLeaf("run"), validLeaf("run")}},
			wantErr: "duplicate subcommand",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateCommand(test.command)
			if err == nil || !bytes.Contains([]byte(err.Error()), []byte(test.wantErr)) {
				t.Fatalf("error = %v, want containing %q", err, test.wantErr)
			}
		})
	}
}

func TestExitCode(t *testing.T) {
	for _, err := range []error{nil, flag.ErrHelp} {
		var stderr bytes.Buffer
		if code := ExitCode(err, &stderr); code != 0 || stderr.Len() != 0 {
			t.Fatalf("ExitCode(%v) = %d, stderr = %q", err, code, stderr.String())
		}
	}

	var stderr bytes.Buffer
	if code := ExitCode(errors.New("failed"), &stderr); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stderr.String() != "error: failed\n" {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestExitOnErrorReturnsForSuccess(t *testing.T) {
	for _, err := range []error{nil, flag.ErrHelp} {
		ExitOnError(err)
	}
}

func TestExecuteRejectsInvalidTree(t *testing.T) {
	err := Execute(Command{Name: "broken"}, nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("has no handler")) {
		t.Fatalf("error = %v", err)
	}
}

func TestPrintCommandList(t *testing.T) {
	var buf bytes.Buffer
	PrintCommandList(&buf, "prosetta generate", "maintainer tools", []Command{
		{Name: "category", Summary: "scaffold a category"},
		{Name: "index", Summary: "compile index.json"},
	})
	out := buf.String()
	want := "Usage:\n  prosetta generate <command> [options]\n\nmaintainer tools\n\nCommands:\n  category  scaffold a category\n  index     compile index.json\n"
	if out != want {
		t.Fatalf("output:\n%q\nwant:\n%q", out, want)
	}
}

func TestIsHelp(t *testing.T) {
	for _, arg := range []string{"help", "-h", "--help"} {
		if !IsHelp(arg) {
			t.Errorf("IsHelp(%q) = false", arg)
		}
	}
	if IsHelp("run") {
		t.Fatal("IsHelp(run) = true")
	}
}
