package execution

import (
	"strings"
	"testing"
	"time"

	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/domain/testcase"
)

func TestPrepareProgramForEveryLanguage(t *testing.T) {
	testCases := []testcase.TestCase{{ID: "1", Input: map[string]any{"name": "World"}, Expected: "Hello World"}}
	for _, lang := range language.AllSupported() {
		t.Run(string(lang), func(t *testing.T) {
			program, err := prepareProgram(lang, "source", testCases, "test-marker", time.Second)
			if err != nil {
				t.Fatalf("prepareProgram: %v", err)
			}
			if program.Entry != "main."+lang.FileExtension() || program.Files["solution."+lang.FileExtension()] != "source" || program.Files[program.Entry] == "" {
				t.Fatalf("invalid generated program = %+v", program)
			}

		})
	}
}

func TestPrepareProgramSupportsStructuredExpectedValues(t *testing.T) {
	for _, lang := range []language.Language{language.Rust, language.Zig} {
		t.Run(string(lang), func(t *testing.T) {
			_, err := prepareProgram(lang, "source", []testcase.TestCase{{
				ID: "1", Expected: map[string]any{"answer": float64(42)},
			}}, "test-marker", time.Second)
			if err != nil {
				t.Fatalf("prepareProgram: %v", err)
			}
		})
	}
}

func TestPrepareGoReportsUnsupportedJSONValue(t *testing.T) {
	_, err := prepareProgram(language.Go, "package main", []testcase.TestCase{{
		ID: "invalid", Input: map[string]any{"channel": make(chan int)},
	}}, "test-marker", time.Second)
	if err == nil || !strings.Contains(err.Error(), "encode input") {
		t.Fatalf("error = %v, want JSON encoding error", err)
	}
}
