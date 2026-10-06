package execution

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"polyglot-rosetta/internal/domain/testcase"
)

// GeneratePythonTestFile creates an executable pytest script inside the workspace using test suite data.
func GeneratePythonTestFile(workspaceDir string, suite *testcase.TestSuite) (string, error) {
	testFilePath := filepath.Join(workspaceDir, "test_solution.py")

	code := `import json
import pytest
from solution import *

`
	for _, tc := range suite.TestCases {
		inputJSON, _ := json.Marshal(tc.Input)
		expectedJSON, _ := json.Marshal(tc.Expected)

		code += fmt.Sprintf(`def test_%s():
    # Description: %s
    raw_input = json.loads(%q)
    expected = json.loads(%q)
    # TODO: invoke user function with raw_input and assert against expected
`, tc.ID, tc.Description, string(inputJSON), string(expectedJSON))
	}

	if err := os.WriteFile(testFilePath, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("failed to generate Python test file: %w", err)
	}

	return testFilePath, nil
}
