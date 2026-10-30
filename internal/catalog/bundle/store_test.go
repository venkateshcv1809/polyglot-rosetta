package bundle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"polyglot-rosetta/internal/catalog"
)

func TestStoreActivationAndReuse(t *testing.T) {
	output := t.TempDir()
	if err := Build("../../../problems", "../../../templates", output); err != nil {
		t.Fatal(err)
	}
	read := func(name string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(output, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	index, templates, archive := read("index.json"), read("templates.tar.gz"), read("fundamentals.tar.gz")
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	var store Store
	if _, err := store.Catalog(); err == nil {
		t.Fatal("uninitialized store accepted")
	}
	if err := store.LoadIndex(index); err != nil {
		t.Fatal(err)
	}
	metadata, _ := store.Catalog()
	if len(metadata.ConceptsByID[id].TestCases) != 0 {
		t.Fatal("index unexpectedly has content")
	}
	archives := map[string][]byte{"fundamentals.tar.gz": archive}
	if err := store.Load(index, templates, archives); err != nil {
		t.Fatal(err)
	}
	active, _ := store.Catalog()
	if active.ConceptsByID[id].Starters["go"] == "" {
		t.Fatal("missing activated starter")
	}
	// Repeated loads and index refreshes reuse the decoded concept maps.
	cached := store.decoded["fundamentals.tar.gz"].concepts
	marker := cached[id]
	marker.Explanation = "reuse marker"
	cached[id] = marker
	if err := store.LoadIndex(index); err != nil {
		t.Fatal(err)
	}
	if err := store.Load(index, templates, archives); err != nil {
		t.Fatal(err)
	}
	active, _ = store.Catalog()
	if active.ConceptsByID[id].Explanation != "reuse marker" {
		t.Fatal("unchanged category was decoded again")
	}
	if err := store.Load(index, templates, map[string][]byte{"fundamentals.tar.gz": {1}}); err == nil {
		t.Fatal("invalid archive accepted")
	}
	after, _ := store.Catalog()
	if after != active {
		t.Fatal("failed activation replaced snapshot")
	}
	if err := store.LoadIndex([]byte(`{}`)); err == nil {
		t.Fatal("invalid index accepted")
	}
	after, _ = store.Catalog()
	if after != active {
		t.Fatal("failed metadata activation replaced snapshot")
	}
	// A new index snapshot must not retain content from the previous release.
	var next catalog.Catalog
	if err := json.Unmarshal(index, &next); err != nil {
		t.Fatal(err)
	}
	next.Title = "Next snapshot"
	changed, _ := json.Marshal(next)
	if err := store.LoadIndex(changed); err != nil {
		t.Fatal(err)
	}
	if len(store.decoded) != 0 {
		t.Fatal("old decoded content retained for a different index")
	}
	after, _ = store.Catalog()
	if after.ConceptsByID[id].Starters != nil {
		t.Fatal("new snapshot contains old content")
	}
}

func TestParseIndexRejectsInvalidGraph(t *testing.T) {
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	valid := catalog.Catalog{ID: "root", Title: "Problems", Categories: []catalog.Category{{ID: "fundamentals", Title: "Fundamentals", Concepts: []catalog.ConceptRef{{ID: id}}}}, ConceptsByID: map[string]catalog.Concept{id: {Path: "fundamentals/hello-world", Title: "Hello", Difficulty: catalog.DifficultyEasy}}}
	for _, mutate := range []func(*catalog.Catalog){
		func(c *catalog.Catalog) { c.ID = "" },
		func(c *catalog.Catalog) { c.Categories[0].ID = "../escape" },
		func(c *catalog.Catalog) {
			c.Categories[0].Concepts = append(c.Categories[0].Concepts, c.Categories[0].Concepts[0])
		},
		func(c *catalog.Catalog) { delete(c.ConceptsByID, id) },
		func(c *catalog.Catalog) { v := c.ConceptsByID[id]; v.Path = "other/path"; c.ConceptsByID[id] = v },
	} {
		data, _ := json.Marshal(valid)
		var candidate catalog.Catalog
		_ = json.Unmarshal(data, &candidate)
		mutate(&candidate)
		data, _ = json.Marshal(candidate)
		if _, err := ParseIndex(data); err == nil {
			t.Fatalf("invalid graph accepted: %s", data)
		}
	}
}
