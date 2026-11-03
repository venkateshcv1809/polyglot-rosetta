package execution

import (
	"context"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/domain/testcase"
)

// Exercise the generated entry points with a real WASI host, independent of any
// browser SDK. The host only supplies compiler targets, argv and WASI imports.
func TestGeneratedProgramsRunUnderWASI(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node is unavailable")
	}
	cases := []testcase.TestCase{
		{ID: "1", Input: map[string]any{"value": map[string]any{"values": []any{true, nil, 3.5}}}, Expected: map[string]any{"values": []any{true, nil, 3.5}}},
		{ID: "2", Input: map[string]any{"value": "Hello 世界 😀\n\"quoted\""}, Expected: "Hello 世界 😀\n\"quoted\""},
	}
	for _, test := range []struct {
		lang     language.Language
		compiler string
		source   string
	}{
		{language.Go, "go", `package main
type Solution struct{}
func (Solution) Solve(input any) any { return input.(map[string]any)["value"] }
`},
		{language.Zig, "zig", `const std = @import("std");
pub const Solution = struct {
    pub fn solve(allocator: std.mem.Allocator, input: std.json.Value) ![]const u8 {
        const value = input.object.get("value").?;
        return if (value == .string) allocator.dupe(u8, value.string) else std.json.Stringify.valueAlloc(allocator, value, .{});
    }
};
`},
	} {
		t.Run(string(test.lang), func(t *testing.T) {
			if _, err := exec.LookPath(test.compiler); err != nil {
				t.Skip("compiler unavailable")
			}
			program, err := prepareProgram(test.lang, test.source, cases, "wasi-test-marker", time.Second)
			if err != nil {
				t.Fatal(err)
			}
			workspace, err := NewWorkspace()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = workspace.Cleanup() })
			for name, source := range program.Files {
				if _, err := workspace.WriteFile(name, source); err != nil {
					t.Fatal(err)
				}
			}
			command := Command{Name: test.compiler, Dir: workspace.Dir}
			if test.lang == language.Go {
				command.Args = []string{"build", "-o", "program.wasm", "."}
				command.Env = []string{"GOWORK=off", "GOOS=wasip1", "GOARCH=wasm"}
			} else {
				command.Args = []string{"build-exe", "main.zig", "-target", "wasm32-wasi", "-O", "ReleaseSafe", "-femit-bin=program.wasm"}
			}
			if result := RunCommand(context.Background(), time.Minute, command); result.Error != nil {
				t.Fatalf("WASI build: %v\n%s", result.Error, result.Stderr)
			}
			launcher, err := workspace.WriteFile("run.cjs", `const fs = require("node:fs");
const { WASI } = require("node:wasi");
const wasi = new WASI({ version: "preview1", args: ["program.wasm", process.argv[2]], returnOnExit: true });
const moduleObject = new WebAssembly.Module(fs.readFileSync("program.wasm"));
const instance = new WebAssembly.Instance(moduleObject, { wasi_snapshot_preview1: wasi.wasiImport });
process.exitCode = wasi.start(instance);
`)
			if err != nil {
				t.Fatal(err)
			}
			for i, testCase := range cases {
				result := RunCommand(context.Background(), 10*time.Second, Command{Name: "node", Args: []string{filepath.Base(launcher), strconv.Itoa(i)}, Dir: workspace.Dir})
				outcome := caseOutcome(testCase, i, result, program.SessionID)
				if !outcome.Passed {
					t.Fatalf("WASI case %d: %+v\n%s\n%s", i, outcome, result.Stdout, result.Stderr)
				}
			}
		})
	}
}
