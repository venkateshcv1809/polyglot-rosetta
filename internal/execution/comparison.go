package execution

import (
	"bytes"
	"encoding/json"
	"math/big"
)

// matchesExpected is the one native comparison rule for every language.
// Rust and Zig retain their plain-text string return interface; other results
// must be valid JSON. Object order is irrelevant, array order and types matter.
func matchesExpected(result caseResult, expected any) bool {
	actualJSON := result.Output
	if result.Format == "text" {
		if _, ok := expected.(string); !ok {
			return false
		}
		actualJSON, _ = json.Marshal(string(result.Output))
	}
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		return false
	}
	actual, ok := decodeJSON(actualJSON)
	if !ok {
		return false
	}
	want, ok := decodeJSON(expectedJSON)
	return ok && equalJSON(actual, want)
}

func decodeJSON(data []byte) (any, bool) {
	if !json.Valid(data) {
		return nil, false
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	err := decoder.Decode(&value)
	return value, err == nil
}

func equalJSON(actual, expected any) bool {
	switch want := expected.(type) {
	case json.Number:
		got, ok := actual.(json.Number)
		if !ok {
			return false
		}
		// Numeric spelling (1 versus 1.0) must not change the result, and
		// converting both sides to float64 would lose large integer differences.
		a, aOK := new(big.Rat).SetString(string(got))
		b, bOK := new(big.Rat).SetString(string(want))
		return aOK && bOK && a.Cmp(b) == 0
	case []any:
		got, ok := actual.([]any)
		if !ok || len(got) != len(want) {
			return false
		}
		for i := range want {
			if !equalJSON(got[i], want[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		got, ok := actual.(map[string]any)
		if !ok || len(got) != len(want) {
			return false
		}
		for key, value := range want {
			item, exists := got[key]
			if !exists || !equalJSON(item, value) {
				return false
			}
		}
		return true
	default:
		// Only null, strings, and booleans reach here after JSON decoding.
		switch actual.(type) {
		case nil, string, bool:
			return actual == expected
		default:
			return false
		}
	}
}
