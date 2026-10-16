package testcase

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// TestCase describes one input and expected result in a concept test suite.
type TestCase struct {
	ID          string         `json:"id"`
	Description string         `json:"description"`
	Input       map[string]any `json:"input"`
	Expected    any            `json:"expected"`
	Hidden      bool           `json:"hidden"`
}

// UnmarshalJSON accepts both string and numeric testcase identifiers. Existing
// catalogs use both forms, while the in-memory representation remains stable.
func (tc *TestCase) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID          json.RawMessage `json:"id"`
		Description string          `json:"description"`
		Input       map[string]any  `json:"input"`
		Expected    any             `json:"expected"`
		Hidden      bool            `json:"hidden"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	id, err := decodeID(raw.ID)
	if err != nil {
		return fmt.Errorf("invalid testcase id: %w", err)
	}
	*tc = TestCase{
		ID:          id,
		Description: raw.Description,
		Input:       raw.Input,
		Expected:    raw.Expected,
		Hidden:      raw.Hidden,
	}
	return nil
}

func decodeID(data json.RawMessage) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("field is required")
	}

	var stringID string
	if err := json.Unmarshal(data, &stringID); err == nil {
		if stringID == "" {
			return "", fmt.Errorf("value cannot be empty")
		}
		return stringID, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return "", err
	}
	number, ok := value.(json.Number)
	if !ok {
		return "", fmt.Errorf("must be a string or number")
	}
	return number.String(), nil
}

// TestSuite groups the test cases associated with a concept.
type TestSuite struct {
	Concept   string     `json:"concept"`
	TestCases []TestCase `json:"test_cases"`
}

// GetTestCases returns either public-only or all test cases depending on the evaluation mode.
func (s *TestSuite) GetTestCases(includeHidden bool) []TestCase {
	if includeHidden {
		return s.TestCases
	}

	publicCases := make([]TestCase, 0)
	for _, tc := range s.TestCases {
		if !tc.Hidden {
			publicCases = append(publicCases, tc)
		}
	}
	return publicCases
}
