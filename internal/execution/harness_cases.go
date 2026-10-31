package execution

import (
	"encoding/json"
	"fmt"

	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/domain/testcase"
)

func encodeCase(testCase testcase.TestCase) (string, string, error) {
	input, err := json.Marshal(testCase.Input)
	if err != nil {
		return "", "", fmt.Errorf("encode input for testcase %q: %w", testCase.ID, err)
	}
	expected, err := json.Marshal(testCase.Expected)
	if err != nil {
		return "", "", fmt.Errorf("encode expected value for testcase %q: %w", testCase.ID, err)
	}
	return string(input), string(expected), nil
}

func expectedText(testCase testcase.TestCase) ([]byte, error) {
	if text, ok := testCase.Expected.(string); ok {
		return []byte(text), nil
	}
	expected, err := json.Marshal(testCase.Expected)
	if err != nil {
		return nil, fmt.Errorf("encode expected value for testcase %q: %w", testCase.ID, err)
	}
	return expected, nil
}

func quoteJSON(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func sourceFilename(lang language.Language) string {
	switch lang {
	case language.Go:
		return "solution.go"
	case language.Python:
		return "solution.py"
	case language.Rust:
		return "solution.rs"
	case language.TypeScript:
		return "solution.ts"
	case language.Zig:
		return "solution.zig"
	default:
		return "solution.txt"
	}
}
