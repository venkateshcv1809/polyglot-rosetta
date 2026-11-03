package run

import (
	"os"

	catalogcontent "polyglot-rosetta/internal/catalog/content"
	"polyglot-rosetta/internal/domain/testcase"
)

// LoadTestCases reads native concept content for the shared run command.
func LoadTestCases(concept Concept) ([]testcase.TestCase, error) {
	suite, err := catalogcontent.LoadTestSuite(os.DirFS(concept.root), concept.Reference)
	if err != nil {
		return nil, err
	}
	return suite.TestCases, nil
}
