package content

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"polyglot-rosetta/internal/catalog"
)

func TestEnrichLocalCatalog(t *testing.T) {
	files := fstest.MapFS{
		"info.json":                                   {Data: []byte(`{"id":"root","title":"Root","items":[{"id":"fundamentals","order":1}]}`)},
		"fundamentals/info.json":                      {Data: []byte(`{"id":"fundamentals","title":"Fundamentals","items":[{"id":"hello-world","order":1}]}`)},
		"fundamentals/hello-world/info.json":          {Data: []byte(`{"id":"hello-world","title":"Hello World","difficulty":"Easy"}`)},
		"fundamentals/hello-world/testcases.json":     {Data: []byte(`[{"id":1,"input":{},"expected":"Hello"},{"id":2,"input":{},"expected":"World"}]`)},
		"fundamentals/hello-world/typescript/main.ts": {Data: []byte("reference solution")},
	}
	common := fstest.MapFS{
		"concept/typescript/template.ts": {Data: []byte("practice starter")},
	}
	index := enrichCatalog(t, files, common)
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	concept, err := index.ResolveConcept(id)
	if err != nil || concept.Path != "fundamentals/hello-world" || len(concept.TestCases) != 2 || concept.Solutions["typescript"] == "" || concept.Starters["typescript"] == "" {
		t.Fatalf("concept = %+v, error = %v", concept, err)
	}
	if _, err := index.ResolveConcept(id[:10]); err != nil {
		t.Fatalf("unique ID prefix: %v", err)
	}
	if concept.Starters["typescript"] == concept.Solutions["typescript"] {
		t.Fatal("starter must not reveal the reference solution")
	}
}

func TestEnrichPrefersConceptStarterOverCommonTemplate(t *testing.T) {
	files := fstest.MapFS{
		"info.json":                          {Data: []byte(`{"id":"root","title":"Root","items":[{"id":"one","order":1}]}`)},
		"one/info.json":                      {Data: []byte(`{"id":"one","title":"One","items":[{"id":"example","order":1}]}`)},
		"one/example/info.json":              {Data: []byte(`{"id":"example","title":"Example","difficulty":"Easy"}`)},
		"one/example/testcases.json":         {Data: []byte(`[{"id":1,"input":{},"expected":true}]`)},
		"one/example/go/template.go.txt":     {Data: []byte("concept starter")},
		"one/example/python/template.py.txt": {Data: []byte("")},
	}
	common := fstest.MapFS{
		"concept/go/template.go":         {Data: []byte("common starter")},
		"concept/python/template.py":     {Data: []byte("python starter")},
		"concept/typescript/template.ts": {Data: []byte("typescript starter")},
	}
	index := enrichCatalog(t, files, common)
	id, _ := catalog.ConceptID("one/example")
	concept := index.ConceptsByID[id]
	python, present := concept.Starters["python"]
	if concept.Starters["go"] != "concept starter" || !present || python != "" || concept.Starters["typescript"] != "typescript starter" {
		t.Fatalf("starters = %+v", concept.Starters)
	}
}

func TestEnrichAllowsRepeatedLeafIDsAtDifferentPaths(t *testing.T) {
	files := fstest.MapFS{
		"info.json":               {Data: []byte(`{"id":"root","title":"Root","items":[{"id":"one","order":1},{"id":"two","order":2}]}`)},
		"one/info.json":           {Data: []byte(`{"id":"one","title":"One","items":[{"id":"same","order":1}]}`)},
		"one/same/info.json":      {Data: []byte(`{"id":"same","title":"Same","difficulty":"Easy"}`)},
		"one/same/testcases.json": {Data: []byte(`[{"id":1,"input":{},"expected":true}]`)},
		"two/info.json":           {Data: []byte(`{"id":"two","title":"Two","items":[{"id":"same","order":1}]}`)},
		"two/same/info.json":      {Data: []byte(`{"id":"same","title":"Same","difficulty":"Easy"}`)},
		"two/same/testcases.json": {Data: []byte(`[{"id":1,"input":{},"expected":true}]`)},
	}
	index := enrichCatalog(t, files, nil)
	for _, ref := range []string{"one/same", "two/same"} {
		id, _ := catalog.ConceptID(ref)
		concept, err := index.ResolveConcept(id)
		if err != nil || concept.Path != ref || concept.Slug != "same" || index.ConceptsByID[id].Path != ref {
			t.Fatalf("ResolveConcept(%q) = %+v, %v", ref, concept, err)
		}
	}
}

func enrichCatalog(t *testing.T, files, templates fs.FS) catalog.Catalog {
	t.Helper()
	metadata, err := catalog.LoadMetadata(files)
	if err != nil {
		t.Fatalf("LoadMetadata: %v", err)
	}
	enriched, err := Enrich(metadata, files, templates)
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	return enriched
}
