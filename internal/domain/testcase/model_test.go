package testcase

import (
	"encoding/json"
	"testing"
)

func TestTestSuiteJSONUnmarshal(t *testing.T) {
	rawJSON := []byte(`{
		"concept": "two-sum",
		"test_cases": [
			{
				"id": "tc_1",
				"description": "basic addition case",
				"input": {"a": 2, "b": 3},
				"expected": 5,
				"hidden": false
			},
			{
				"id": "tc_2",
				"description": "edge case with negative numbers",
				"input": {"a": -1, "b": 1},
				"expected": 0,
				"hidden": true
			}
		]
	}`)

	var suite TestSuite
	if err := json.Unmarshal(rawJSON, &suite); err != nil {
		t.Fatalf("Failed to unmarshal test suite JSON: %v", err)
	}

	if suite.Concept != "two-sum" {
		t.Errorf("Expected concept 'two-sum', got '%s'", suite.Concept)
	}

	if len(suite.TestCases) != 2 {
		t.Fatalf("Expected 2 test cases, got %d", len(suite.TestCases))
	}

	// Validate first test case
	tc1 := suite.TestCases[0]
	if tc1.ID != "tc_1" || tc1.Hidden != false {
		t.Errorf("Test case 1 metadata incorrect: %+v", tc1)
	}
	if val, ok := tc1.Input["a"].(float64); !ok || val != 2 {
		t.Errorf("Test case 1 input mapping failed: %+v", tc1.Input)
	}

	// Validate second test case (hidden)
	tc2 := suite.TestCases[1]
	if tc2.ID != "tc_2" || !tc2.Hidden {
		t.Errorf("Test case 2 hidden flag incorrect: %+v", tc2)
	}
}
