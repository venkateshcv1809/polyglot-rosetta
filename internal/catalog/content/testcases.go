package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"

	"polyglot-rosetta/internal/domain/testcase"
)

// LoadTestSuite reads a concept's test cases from a catalog filesystem.
// Both array and legacy wrapped-suite files are accepted.
func LoadTestSuite(files fs.FS, directory string) (*testcase.TestSuite, error) {
	name := path.Join(directory, "testcases.json")
	data, err := fs.ReadFile(files, name)
	if err != nil {
		return nil, fmt.Errorf("read testcases for %q: %w", directory, err)
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("testcases for %q are empty", directory)
	}
	suite := &testcase.TestSuite{}
	if data[0] == '[' {
		if err := json.Unmarshal(data, &suite.TestCases); err != nil {
			return nil, fmt.Errorf("decode testcases for %q: %w", directory, err)
		}
	} else if err := json.Unmarshal(data, suite); err != nil {
		return nil, fmt.Errorf("decode testcases for %q: %w", directory, err)
	}
	if suite.Concept == "" {
		suite.Concept = path.Base(directory)
	}
	if len(suite.TestCases) == 0 {
		return nil, fmt.Errorf("concept %q has no testcases", directory)
	}
	ids := make(map[string]struct{}, len(suite.TestCases))
	for _, test := range suite.TestCases {
		if _, exists := ids[test.ID]; exists {
			return nil, fmt.Errorf("concept %q has duplicate testcase ID %q", directory, test.ID)
		}
		ids[test.ID] = struct{}{}
	}
	return suite, nil
}
