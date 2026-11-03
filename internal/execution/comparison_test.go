package execution

import (
	"encoding/json"
	"testing"
)

func TestMatchesExpected(t *testing.T) {
	for _, test := range []struct {
		name, output, format string
		expected             any
		matches              bool
	}{
		{"boolean", "true", "json", true, true},
		{"number is not boolean", "1", "json", true, false},
		{"nested types", `{"values":[1]}`, "json", map[string]any{"values": []any{true}}, false},
		{"object ordering", `{"b":2,"a":1}`, "json", map[string]any{"a": 1, "b": 2}, true},
		{"array ordering", `[2,1]`, "json", []any{1, 2}, false},
		{"array length", `[null]`, "json", []any{}, false},
		{"null is not empty array", "null", "json", []any{}, false},
		{"missing null field", `{}`, "json", map[string]any{"a": nil}, false},
		{"equivalent numbers", "1e0", "json", 1, true},
		{"distinct large numbers", "9007199254740993", "json", json.Number("9007199254740992"), false},
		{"invalid leading zero", "01", "json", 1, false},
		{"invalid trailing decimal", "1.", "json", 1, false},
		{"multiple values", "1 1", "json", 1, false},
		{"escaped Unicode", `["\ud83d\ude00"]`, "json", []any{"😀"}, true},
		{"plain multiline string", "Hello\n世界\x00", "text", "Hello\n世界\x00", true},
		{"plain empty string", "", "text", "", true},
		{"text is not number", "1", "text", 1, false},
		{"JSON string", `"Hello\n世界"`, "json", "Hello\n世界", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := matchesExpected(caseResult{Output: []byte(test.output), Format: test.format}, test.expected)
			if got != test.matches {
				t.Fatalf("matchesExpected(%q, %#v) = %v", test.output, test.expected, got)
			}
		})
	}
}
