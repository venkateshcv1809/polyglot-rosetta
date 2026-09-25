package execution

import (
	"os"
	"testing"

	"polyglot-rosetta/internal/domain/testcase"
)

func TestGeneratePythonTestFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "py_generator_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	suite := &testcase.TestSuite{
		Concept: "two-sum",
		TestCases: []testcase.TestCase{
			{
				ID:          "tc_1",
				Description: "python basic case",
				Input:       map[string]any{"nums": []int{2, 7, 11, 15}, "target": 9},
				Expected:    []int{0, 1},
				Hidden:      false,
			},
		},
	}

	path, err := GeneratePythonTestFile(tmpDir, suite)
	if err != nil {
		t.Fatalf("unexpected error generating python test file: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("generated python test file does not exist on disk: %v", err)
	}
}
