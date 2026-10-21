package execution

import (
	"os"
	"strings"
	"testing"

	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/domain/testcase"
)

func TestPrepareProgramForEveryLanguage(t *testing.T) {
	testCases := []testcase.TestCase{{ID: "1", Input: map[string]any{"name": "World"}, Expected: "Hello World"}}
	for _, lang := range language.AllSupported() {
		t.Run(string(lang), func(t *testing.T) {
			workspace, err := NewWorkspace()
			if err != nil {
				t.Fatalf("NewWorkspace: %v", err)
			}
			t.Cleanup(func() { _ = workspace.Cleanup() })

			program, err := prepareProgram(workspace, lang, "source", testCases, "test-marker")
			if err != nil {
				t.Fatalf("prepareProgram: %v", err)
			}
			if program.run.Name == "" || program.run.Dir != workspace.Dir {
				t.Fatalf("run command = %+v", program.run)
			}
			entries, err := os.ReadDir(workspace.Dir)
			if err != nil || len(entries) < 2 {
				t.Fatalf("workspace entries = %v, error = %v", entries, err)
			}
		})
	}
}

func TestPrepareProgramSupportsStructuredExpectedValues(t *testing.T) {
	for _, lang := range []language.Language{language.Rust, language.Zig} {
		t.Run(string(lang), func(t *testing.T) {
			workspace, err := NewWorkspace()
			if err != nil {
				t.Fatalf("NewWorkspace: %v", err)
			}
			t.Cleanup(func() { _ = workspace.Cleanup() })

			_, err = prepareProgram(workspace, lang, "source", []testcase.TestCase{{
				ID: "1", Expected: map[string]any{"answer": float64(42)},
			}}, "test-marker")
			if err != nil {
				t.Fatalf("prepareProgram: %v", err)
			}
		})
	}
}

func TestPrepareGoReportsUnsupportedJSONValue(t *testing.T) {
	workspace, err := NewWorkspace()
	if err != nil {
		t.Fatalf("NewWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = workspace.Cleanup() })

	_, err = prepareProgram(workspace, language.Go, "package main", []testcase.TestCase{{
		ID: "invalid", Input: map[string]any{"channel": make(chan int)},
	}}, "test-marker")
	if err == nil || !strings.Contains(err.Error(), "encode input") {
		t.Fatalf("error = %v, want JSON encoding error", err)
	}
}

func TestNativeHarnessesParseBeforeTiming(t *testing.T) {
	testCases := []testcase.TestCase{{ID: "1", Input: map[string]any{"name": "World"}, Expected: "Hello World"}}
	for _, test := range []struct {
		name      string
		prepare   func([]testcase.TestCase, string) (map[string]string, preparedProgram, error)
		file      string
		parseCall string
		timerCall string
		buildFlag string
	}{
		{"rust", prepareRust, "main.rs", "JsonParser::new(input_text).parse()", "let started = Instant::now()", "opt-level=3"},
		{"zig", prepareZig, "main.zig", "std.json.parseFromSlice(std.json.Value", "const started_0 = monotonicNs();", "ReleaseFast"},
	} {
		t.Run(test.name, func(t *testing.T) {
			files, program, err := test.prepare(testCases, "test-marker")
			if err != nil {
				t.Fatalf("prepare: %v", err)
			}
			source := files[test.file]
			parseAt, timerAt := strings.LastIndex(source, test.parseCall), strings.LastIndex(source, test.timerCall)
			if parseAt < 0 || timerAt <= parseAt {
				t.Fatalf("input must be parsed before timing in %s", test.file)
			}
			if len(program.build) != 1 || !strings.Contains(strings.Join(program.build[0].Args, " "), test.buildFlag) {
				t.Fatalf("optimized build flag missing: %+v", program.build)
			}
		})
	}
}
