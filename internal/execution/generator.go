package execution

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"polyglot-rosetta/internal/domain/testcase"
)

// GenerateGoTestFile creates a compilable _test.go file inside the workspace using test suite data.
func GenerateGoTestFile(workspaceDir string, suite *testcase.TestSuite) (string, error) {
	testFilePath := filepath.Join(workspaceDir, "solution_test.go")

	code := `package main

import (
	"encoding/json"
	"testing"
)

func TestMainExecution(t *testing.T) {
`
	for _, tc := range suite.TestCases {
		inputJSON, _ := json.Marshal(tc.Input)
		expectedJSON, _ := json.Marshal(tc.Expected)

		code += fmt.Sprintf(`	t.Run(%q, func(t *testing.T) {
		// Testcase ID: %s
		// Description: %s
		// Input payload: %s
		// Expected output: %s
		// (Test harness evaluates user code here)
	})
`, tc.ID, tc.ID, tc.Description, string(inputJSON), string(expectedJSON))
	}

	code += "}\n"

	if err := os.WriteFile(testFilePath, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("failed to generate Go test file: %w", err)
	}

	return testFilePath, nil
}
