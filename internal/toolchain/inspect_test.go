package toolchain

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestInspectBinaryUsesVersionArguments(t *testing.T) {
	path := writeTestTool(t, "printf '%s\\nextra line\\n' \"$1\"")
	for _, test := range []struct {
		binary string
		want   string
	}{
		{binary: "go", want: "version"},
		{binary: "zig", want: "version"},
		{binary: "python3", want: "--version"},
		{binary: "rustc", want: "--version"},
		{binary: "tsc", want: "--version"},
		{binary: "node", want: "--version"},
	} {
		t.Run(test.binary, func(t *testing.T) {
			version, err := InspectBinary(test.binary, path)
			if err != nil || version != test.want {
				t.Fatalf("InspectBinary(%q) = %q, %v; want %q", test.binary, version, err, test.want)
			}
		})
	}
}

func TestInspectBinaryRejectsEmptyOutput(t *testing.T) {
	path := writeTestTool(t, "printf ' \\n'")
	if _, err := InspectBinary("go", path); err == nil || !strings.Contains(err.Error(), "empty version") {
		t.Fatalf("error = %v, want empty version", err)
	}
}

func TestInspectBinaryReportsExecutionFailure(t *testing.T) {
	path := writeTestTool(t, "exit 7")
	_, err := InspectBinary("go", path)
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 7 {
		t.Fatalf("error = %v, want exit status 7", err)
	}
}

func TestInspectBinaryTimesOut(t *testing.T) {
	path := writeTestTool(t, "while :; do :; done")
	_, err := inspectBinary("go", path, 20*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want deadline exceeded", err)
	}
}

func writeTestTool(t *testing.T, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell fixtures require a Unix host")
	}
	path := filepath.Join(t.TempDir(), "test-tool")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatalf("write test tool: %v", err)
	}
	return path
}
