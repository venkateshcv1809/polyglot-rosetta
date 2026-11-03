package execution

import (
	"context"
	"os/exec"
	"testing"

	"polyglot-rosetta/internal/domain/evaluation"
	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/domain/testcase"
)

func TestEvaluateComparisonRegressions(t *testing.T) {
	for _, test := range []struct {
		name     string
		lang     language.Language
		source   string
		expected any
		passes   bool
	}{
		{"Python number cannot stand in for boolean", language.Python, "class Solution:\n    def solve(self, data):\n        return 1\n", true, false},
		{"Python nested boolean retains its type", language.Python, "class Solution:\n    def solve(self, data):\n        return {'values': [1]}\n", map[string]any{"values": []any{true}}, false},
		{"TypeScript sparse array is not empty", language.TypeScript, "export class Solution { solve(_: unknown): unknown { return new Array(1); } }", []any{}, false},
		{"TypeScript undefined element is not empty", language.TypeScript, "export class Solution { solve(_: unknown): unknown { return [undefined]; } }", []any{}, false},
		{"Rust valid surrogate pair", language.Rust, `pub struct Solution; impl Solution { pub fn solve(_: &crate::JsonValue)->String { r#"["\ud83d\ude00"]"#.to_string() } }`, []any{"😀"}, true},
		{"Rust invalid JSON number", language.Rust, `pub struct Solution; impl Solution { pub fn solve(_: &crate::JsonValue)->String { "01".to_string() } }`, 1, false},
		{"Rust formatted multiline JSON", language.Rust, "pub struct Solution; impl Solution { pub fn solve(_: &crate::JsonValue)->String { \"[\\n1,\\n2\\n]\".to_string() } }", []any{1, 2}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := exec.LookPath(map[language.Language]string{language.Python: "python3", language.Rust: "rustc", language.TypeScript: "tsc"}[test.lang]); err != nil {
				t.Skip("language toolchain is unavailable")
			}
			suite := &testcase.TestSuite{Concept: "comparison", TestCases: []testcase.TestCase{{ID: "1", Input: map[string]any{}, Expected: test.expected}}}
			got, err := EvaluateSuite(context.Background(), evaluation.Request{Language: test.lang, SourceCode: test.source}, suite)
			if err != nil {
				t.Fatal(err)
			}
			if got.Report.PassedAll != test.passes {
				t.Fatalf("report = %+v", got.Report)
			}
		})
	}
}

func TestEvaluateRustConstructsNestedInput(t *testing.T) {
	if _, err := exec.LookPath("rustc"); err != nil {
		t.Skip("Rust toolchain is unavailable")
	}
	source := `use crate::JsonValue;
pub struct Solution;
impl Solution {
    pub fn solve(input: &JsonValue) -> String {
        match input {
            JsonValue::Object(fields) => match fields.get("values") {
                Some(JsonValue::Array(values)) if values == &vec![
                    JsonValue::Null, JsonValue::Bool(true), JsonValue::Number(-1.25e20),
                    JsonValue::String("line\n😀\0".to_string()), JsonValue::Array(vec![]),
                    JsonValue::Object(std::collections::BTreeMap::new())
                ] => "ok".to_string(),
                _ => "wrong".to_string(),
            },
            _ => "wrong".to_string(),
        }
    }
}`
	suite := &testcase.TestSuite{Concept: "nested", TestCases: []testcase.TestCase{{
		ID: "1", Input: map[string]any{"values": []any{nil, true, -1.25e20, "line\n😀\x00", []any{}, map[string]any{}}}, Expected: "ok",
	}}}
	got, err := EvaluateSuite(context.Background(), evaluation.Request{Language: language.Rust, SourceCode: source}, suite)
	if err != nil || !got.Report.PassedAll {
		t.Fatalf("report = %+v, error = %v", got.Report, err)
	}
}
