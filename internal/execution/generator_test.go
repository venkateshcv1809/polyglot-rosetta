package execution

import (
	"os"
	"testing"

	"polyglot-rosetta/internal/domain/testcase"
)

func TestGenerateGoTestFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	suite := &testcase.TestSuite{
		Concept: "two-sum",
		TestCases: []testcase.TestCase{
			{
				ID:          "tc_1",
				Description: "basic test case",
				Input:       map[string]any{"a": 1, "b": 2},
				Expected:    3,
				Hidden:      false,
			},
		},
	}

	path, err := GenerateGoTestFile(tmpDir, suite)
	if err != nil {
		t.Fatalf("unexpected error generating test file: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("generated test file does not exist on disk: %v", err)
	}
}
