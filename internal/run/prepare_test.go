package run

import (
	"errors"
	"os"
	"strings"
	"testing"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/catalog"
)

func TestPrepareChoosesCatalogOrExplicitSource(t *testing.T) {
	t.Setenv(catalogEnvironment, writePracticeCatalog(t))
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	for _, test := range []struct {
		name   string
		source *string
		want   string
	}{
		{name: "catalog", want: "Solve(any)"},
		{name: "explicit", source: testString("custom source"), want: "custom source"},
	} {
		t.Run(test.name, func(t *testing.T) {
			request, concept, err := Prepare(id[:10], "go", runapi.ModePublic, test.source)
			if err != nil {
				t.Fatal(err)
			}
			if concept.ID != id || request.ID != id || !strings.Contains(request.Source, test.want) {
				t.Fatalf("request = %+v, concept = %+v", request, concept)
			}
		})
	}
}

func TestPreparePreservesConceptWhenSolutionIsMissing(t *testing.T) {
	t.Setenv(catalogEnvironment, writePracticeCatalog(t))
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	_, concept, err := Prepare(id, "python", runapi.ModePublic, nil)
	if !errors.Is(err, os.ErrNotExist) || concept.ID != id {
		t.Fatalf("concept = %+v, error = %v", concept, err)
	}
}

func testString(value string) *string { return &value }
