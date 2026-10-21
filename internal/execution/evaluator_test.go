package execution

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	domainexecution "polyglot-rosetta/internal/domain/execution"
	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/domain/testcase"
)

func TestEvaluateGoFiltersHiddenCases(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain is unavailable")
	}
	suite := evaluationSuite()
	source := `package main
type Solution struct{}
func (Solution) Solve(input any) any { return "Hello " + input.(map[string]any)["name"].(string) }
`

	evaluation, err := EvaluateSuite(context.Background(), domainexecution.Request{
		Language: language.Go, SourceCode: source,
	}, suite)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !evaluation.Report.PassedAll || evaluation.Report.TotalTests != 1 {
		t.Fatalf("report = %+v", evaluation.Report)
	}
	if evaluation.BuildDuration == nil || *evaluation.BuildDuration <= 0 || *evaluation.BuildDuration >= evaluation.Duration {
		t.Fatalf("build duration = %v, elapsed = %v", evaluation.BuildDuration, evaluation.Duration)
	}
}

func TestEvaluateSuiteUsesInMemoryTestCases(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python toolchain is unavailable")
	}
	suite := &testcase.TestSuite{Concept: "embedded", TestCases: []testcase.TestCase{{
		ID: "one", Input: map[string]any{"name": "World"}, Expected: "Hello World",
	}}}
	evaluation, err := EvaluateSuite(context.Background(), domainexecution.Request{
		Language:   language.Python,
		SourceCode: "class Solution:\n    def solve(self, input_data):\n        return 'Hello ' + input_data['name']\n",
	}, suite)
	if err != nil {
		t.Fatalf("EvaluateSuite: %v", err)
	}
	if !evaluation.Report.PassedAll || evaluation.Report.Concept != "embedded" {
		t.Fatalf("report = %+v", evaluation.Report)
	}
	if evaluation.BuildDuration != nil {
		t.Fatalf("Python should not have a build duration: %v", *evaluation.BuildDuration)
	}
}

func TestEvaluatePythonIncludesHiddenCases(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python toolchain is unavailable")
	}
	suite := evaluationSuite()
	source := "class Solution:\n    def solve(self, input_data):\n        return 'Hello ' + input_data['name']\n"

	evaluation, err := EvaluateSuite(context.Background(), domainexecution.Request{
		Language: language.Python, SourceCode: source, IsSubmit: true,
	}, suite)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !evaluation.Report.PassedAll || evaluation.Report.TotalTests != 2 {
		t.Fatalf("report = %+v", evaluation.Report)
	}
}

func TestEvaluateReportsPerTestSolutionTime(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python toolchain is unavailable")
	}
	suite := &testcase.TestSuite{Concept: "timing", TestCases: []testcase.TestCase{
		{ID: "slow", Input: map[string]any{"value": "slow"}, Expected: "slow"},
		{ID: "fast", Input: map[string]any{"value": "fast"}, Expected: "fast"},
	}}
	evaluation, err := EvaluateSuite(context.Background(), domainexecution.Request{
		Language:   language.Python,
		SourceCode: "import time\nclass Solution:\n    def solve(self, data):\n        if data['value'] == 'slow':\n            time.sleep(0.1)\n        return data['value']\n",
	}, suite)
	if err != nil {
		t.Fatalf("EvaluateSuite: %v", err)
	}
	outcomes := evaluation.Report.Outcomes
	if len(outcomes) != 2 || outcomes[0].TestCaseID != "slow" || outcomes[1].TestCaseID != "fast" {
		t.Fatalf("outcomes = %+v", outcomes)
	}
	if outcomes[0].SolutionTimeNs < int64(100*time.Millisecond) || outcomes[1].SolutionTimeNs >= int64(50*time.Millisecond) {
		t.Fatalf("per-test solution times = %+v", outcomes)
	}
}

func TestEvaluateIsolatesEachCase(t *testing.T) {
	suite := &testcase.TestSuite{Concept: "isolated", TestCases: []testcase.TestCase{
		{ID: "first", Input: map[string]any{}, Expected: "one"},
		{ID: "second", Input: map[string]any{}, Expected: "one"},
	}}
	for _, test := range []struct {
		name       string
		language   language.Language
		binaries   []string
		sourceCode string
	}{
		{
			name: "go", language: language.Go, binaries: []string{"go"},
			sourceCode: "package main\nvar calls int\ntype Solution struct{}\nfunc (Solution) Solve(any) any { calls++; if calls == 1 { return \"one\" }; return \"again\" }\n",
		},
		{
			name: "python", language: language.Python, binaries: []string{"python3"},
			sourceCode: "calls = 0\nclass Solution:\n    def solve(self, data):\n        global calls\n        calls += 1\n        return 'one' if calls == 1 else 'again'\n",
		},
		{
			name: "typescript", language: language.TypeScript, binaries: []string{"tsc", "node"},
			sourceCode: "let calls = 0; export class Solution { solve(input: unknown): unknown { calls++; return calls === 1 ? 'one' : 'again'; } }\n",
		},
		{
			name: "rust", language: language.Rust, binaries: []string{"rustc"},
			sourceCode: "use crate::JsonValue;\nuse std::sync::atomic::{AtomicUsize, Ordering};\nstatic CALLS: AtomicUsize = AtomicUsize::new(0);\npub struct Solution;\nimpl Solution { pub fn solve(_: &JsonValue) -> String { if CALLS.fetch_add(1, Ordering::SeqCst) == 0 { \"one\" } else { \"again\" }.to_string() } }\n",
		},
		{
			name: "zig", language: language.Zig, binaries: []string{"zig"},
			sourceCode: "const std = @import(\"std\");\nvar calls: usize = 0;\npub const Solution = struct { pub fn solve(allocator: std.mem.Allocator, _: std.json.Value) ![]const u8 { calls += 1; return allocator.dupe(u8, if (calls == 1) \"one\" else \"again\"); } };\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, binary := range test.binaries {
				if _, err := exec.LookPath(binary); err != nil {
					t.Skipf("%s is unavailable", binary)
				}
			}
			evaluation, err := EvaluateSuite(context.Background(), domainexecution.Request{
				Language: test.language, SourceCode: test.sourceCode,
			}, suite)
			if err != nil {
				t.Fatalf("EvaluateSuite: %v", err)
			}
			if !evaluation.Report.PassedAll || len(evaluation.Report.Outcomes) != 2 ||
				evaluation.Report.Outcomes[0].Status != domainexecution.TestPassed ||
				evaluation.Report.Outcomes[1].Status != domainexecution.TestPassed {
				t.Fatalf("report = %+v", evaluation.Report)
			}
		})
	}
}

func TestEvaluateContinuesAfterCrashForEveryLanguage(t *testing.T) {
	suite := &testcase.TestSuite{Concept: "crash", TestCases: []testcase.TestCase{
		{ID: "crash", Input: map[string]any{"crash": true}, Expected: "ok"},
		{ID: "next", Input: map[string]any{"crash": false}, Expected: "ok"},
	}}
	for _, test := range []struct {
		name       string
		language   language.Language
		binaries   []string
		sourceCode string
	}{
		{
			name: "go", language: language.Go, binaries: []string{"go"},
			sourceCode: "package main\ntype Solution struct{}\nfunc (Solution) Solve(input any) any { if input.(map[string]any)[\"crash\"] == true { panic(\"boom\") }; return \"ok\" }\n",
		},
		{
			name: "python", language: language.Python, binaries: []string{"python3"},
			sourceCode: "class Solution:\n    def solve(self, data):\n        if data['crash']: raise RuntimeError('boom')\n        return 'ok'\n",
		},
		{
			name: "typescript", language: language.TypeScript, binaries: []string{"tsc", "node"},
			sourceCode: "export class Solution { solve(input: any): string { if (input.crash) throw new Error('boom'); return 'ok'; } }\n",
		},
		{
			name: "rust", language: language.Rust, binaries: []string{"rustc"},
			sourceCode: "use crate::JsonValue;\npub struct Solution;\nimpl Solution { pub fn solve(input: &JsonValue) -> String { if matches!(input, JsonValue::Object(fields) if matches!(fields.get(\"crash\"), Some(JsonValue::Bool(true)))) { panic!(\"boom\") }; \"ok\".to_string() } }\n",
		},
		{
			name: "zig", language: language.Zig, binaries: []string{"zig"},
			sourceCode: "const std = @import(\"std\");\npub const Solution = struct { pub fn solve(allocator: std.mem.Allocator, input: std.json.Value) ![]const u8 { if (input.object.get(\"crash\").? == .bool and input.object.get(\"crash\").?.bool) return error.Boom; return allocator.dupe(u8, \"ok\"); } };\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, binary := range test.binaries {
				if _, err := exec.LookPath(binary); err != nil {
					t.Skipf("%s is unavailable", binary)
				}
			}
			evaluation, err := EvaluateSuite(context.Background(), domainexecution.Request{
				Language: test.language, SourceCode: test.sourceCode, IsSubmit: true,
			}, suite)
			if err != nil {
				t.Fatalf("EvaluateSuite: %v", err)
			}
			outcomes := evaluation.Report.Outcomes
			if evaluation.Report.PassedAll || evaluation.Report.PassedTests != 1 || len(outcomes) != 2 ||
				outcomes[0].Status != domainexecution.TestCrashed || outcomes[0].Message == "" ||
				outcomes[1].Status != domainexecution.TestPassed || outcomes[1].SolutionTimeNs <= 0 {
				t.Fatalf("report = %+v", evaluation.Report)
			}
		})
	}
}

func TestEvaluateContinuesAfterWrongAnswer(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain is unavailable")
	}
	suite := &testcase.TestSuite{Concept: "wrong-answer", TestCases: []testcase.TestCase{
		{ID: "wrong", Input: map[string]any{"wrong": true}, Expected: "ok"},
		{ID: "next", Input: map[string]any{"wrong": false}, Expected: "ok"},
	}}
	evaluation, err := EvaluateSuite(context.Background(), domainexecution.Request{
		Language: language.Go, SourceCode: "package main\ntype Solution struct{}\nfunc (Solution) Solve(input any) any { if input.(map[string]any)[\"wrong\"] == true { return \"wrong\" }; return \"ok\" }\n",
	}, suite)
	if err != nil {
		t.Fatalf("EvaluateSuite: %v", err)
	}
	if evaluation.Report.PassedTests != 1 ||
		evaluation.Report.Outcomes[0].Status != domainexecution.TestWrongAnswer ||
		evaluation.Report.Outcomes[1].Status != domainexecution.TestPassed {
		t.Fatalf("report = %+v", evaluation.Report)
	}
}

func TestEvaluateContinuesAfterEarlyExit(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain is unavailable")
	}
	suite := &testcase.TestSuite{Concept: "early-exit", TestCases: []testcase.TestCase{
		{ID: "exit", Input: map[string]any{"exit": true}, Expected: "ok"},
		{ID: "next", Input: map[string]any{"exit": false}, Expected: "ok"},
	}}
	evaluation, err := EvaluateSuite(context.Background(), domainexecution.Request{
		Language:   language.Go,
		SourceCode: "package main\nimport \"os\"\ntype Solution struct{}\nfunc (Solution) Solve(input any) any { if input.(map[string]any)[\"exit\"] == true { os.Exit(0) }; return \"ok\" }\n",
	}, suite)
	if err != nil {
		t.Fatalf("EvaluateSuite: %v", err)
	}
	if evaluation.Report.PassedTests != 1 ||
		evaluation.Report.Outcomes[0].Status != domainexecution.TestCrashed ||
		evaluation.Report.Outcomes[1].Status != domainexecution.TestPassed {
		t.Fatalf("report = %+v", evaluation.Report)
	}
}

func TestEvaluateStructuredResultsForEveryLanguage(t *testing.T) {
	suite := structuredEvaluationSuite()
	tests := []struct {
		name       string
		language   language.Language
		binaries   []string
		sourceCode string
	}{
		{
			name: "go", language: language.Go, binaries: []string{"go"},
			sourceCode: "package main\ntype Solution struct{}\nfunc (Solution) Solve(any) any { return map[string]any{\"label\": \"ok\", \"answer\": float64(42)} }\n",
		},
		{
			name: "python", language: language.Python, binaries: []string{"python3"},
			sourceCode: "class Solution:\n    def solve(self, input_data):\n        return {\"label\": \"ok\", \"answer\": 42}\n",
		},
		{
			name: "typescript", language: language.TypeScript, binaries: []string{"tsc", "node"},
			sourceCode: "export class Solution { solve(input: unknown): unknown { return { label: \"ok\", answer: 42 }; } }\n",
		},
		{
			name: "rust", language: language.Rust, binaries: []string{"rustc"},
			sourceCode: `pub struct Solution;
impl Solution { pub fn solve(_: &crate::JsonValue) -> String { r#"{ "label": "ok", "answer": 42 }"#.to_string() } }
`,
		},
		{
			name: "zig", language: language.Zig, binaries: []string{"zig"},
			sourceCode: `const std = @import("std");
pub const Solution = struct { pub fn solve(allocator: std.mem.Allocator, _: std.json.Value) ![]const u8 { return allocator.dupe(u8, "{ \"label\": \"ok\", \"answer\": 42 }"); } };
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, binary := range test.binaries {
				if _, err := exec.LookPath(binary); err != nil {
					t.Skipf("%s is unavailable", binary)
				}
			}
			evaluation, err := EvaluateSuite(context.Background(), domainexecution.Request{
				Language: test.language, SourceCode: test.sourceCode,
			}, suite)
			if err != nil {
				t.Fatalf("Evaluate: %v", err)
			}
			if !evaluation.Report.PassedAll {
				t.Fatalf("report = %+v", evaluation.Report)
			}
			if len(evaluation.Report.Outcomes) != 1 || evaluation.Report.Outcomes[0].SolutionTimeNs <= 0 {
				t.Fatalf("solution time missing: %+v", evaluation.Report.Outcomes)
			}
		})
	}
}

func TestEvaluateReturnsFailedReportForWrongAnswer(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain is unavailable")
	}
	evaluation, err := EvaluateSuite(context.Background(), domainexecution.Request{
		Language:   language.Go,
		SourceCode: "package main\ntype Solution struct{}\nfunc (Solution) Solve(any) any { return \"wrong\" }\n",
	}, evaluationSuite())
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if evaluation.Report.PassedAll || evaluation.Report.PassedTests != 0 {
		t.Fatalf("report = %+v", evaluation.Report)
	}
}

func TestEvaluateReportsCompileFailure(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain is unavailable")
	}
	_, err := EvaluateSuite(context.Background(), domainexecution.Request{
		Language: language.Go, SourceCode: "not go source",
	}, evaluationSuite())
	if err == nil || !strings.Contains(err.Error(), `build command "go" failed`) {
		t.Fatalf("error = %v", err)
	}
}

func TestEvaluateRejectsNonPositiveTimeout(t *testing.T) {
	_, err := EvaluateSuiteWithTimeout(context.Background(), domainexecution.Request{}, evaluationSuite(), 0)
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("error = %v", err)
	}
}

func evaluationSuite() *testcase.TestSuite {
	return &testcase.TestSuite{Concept: "hello-world", TestCases: []testcase.TestCase{
		{ID: "1", Input: map[string]any{"name": "World"}, Expected: "Hello World"},
		{ID: "2", Input: map[string]any{"name": "Rosetta"}, Expected: "Hello Rosetta", Hidden: true},
	}}
}

func structuredEvaluationSuite() *testcase.TestSuite {
	return &testcase.TestSuite{Concept: "structured-output", TestCases: []testcase.TestCase{{
		ID: "structured", Input: map[string]any{},
		Expected: map[string]any{"answer": float64(42), "label": "ok"},
	}}}
}

func TestEvaluateTimesOut(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python toolchain is unavailable")
	}
	suite := &testcase.TestSuite{Concept: "timeout", TestCases: []testcase.TestCase{
		{ID: "slow", Input: map[string]any{"slow": true}, Expected: "ok"},
		{ID: "next", Input: map[string]any{"slow": false}, Expected: "ok"},
	}}
	evaluation, err := EvaluateSuiteWithTimeout(context.Background(), domainexecution.Request{
		Language:   language.Python,
		SourceCode: "import time\nclass Solution:\n    def solve(self, data):\n        if data['slow']: time.sleep(0.3)\n        return 'ok'\n",
	}, suite, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("EvaluateSuiteWithTimeout: %v", err)
	}
	if evaluation.Report.PassedAll || evaluation.Report.Outcomes[0].Status != domainexecution.TestTimeout ||
		evaluation.Report.Outcomes[1].Status != domainexecution.TestPassed {
		t.Fatalf("report = %+v", evaluation.Report)
	}
}
