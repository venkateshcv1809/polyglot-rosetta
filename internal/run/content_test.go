package run

import (
	"os"
	"path/filepath"
	"testing"

	"polyglot-rosetta/internal/catalog"
)

func TestLoadTestCasesFromResolvedConcept(t *testing.T) {
	root := writePracticeCatalog(t)
	t.Setenv(catalogEnvironment, root)
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	concept, err := ResolveConcept(id)
	if err != nil {
		t.Fatal(err)
	}
	cases, err := LoadTestCases(concept)
	if err != nil || len(cases) != 1 || cases[0].Expected != true {
		t.Fatalf("cases = %+v, error = %v", cases, err)
	}
	if err := os.WriteFile(filepath.Join(root, "fundamentals", "hello-world", "testcases.json"), []byte("invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadTestCases(concept); err == nil {
		t.Fatal("invalid testcase content accepted")
	}
}
