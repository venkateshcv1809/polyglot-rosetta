package catalogcmd

import (
	"context"
	"errors"
	"testing"

	catalogapi "polyglot-rosetta/internal/api/catalog"
	"polyglot-rosetta/internal/catalog"
)

type fakeHost struct {
	root    string
	catalog *catalog.Catalog
	err     error
}

func (h fakeHost) Activate(context.Context) (string, *catalog.Catalog, error) {
	return h.root, h.catalog, h.err
}
func TestCatalogSharedResult(t *testing.T) {
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	loaded := &catalog.Catalog{ConceptsByID: map[string]catalog.Concept{id: {Path: "fundamentals/hello-world"}}}
	for _, host := range []fakeHost{{root: "/problems"}, {catalog: loaded}} {
		result, err := Catalog(context.Background(), host)
		if err != nil || result.Status != catalogapi.StatusOK || result.Root != host.root || result.Catalog != host.catalog {
			t.Fatalf("result=%+v err=%v", result, err)
		}
	}
	sentinel := errors.New("load failed")
	result, err := Catalog(context.Background(), fakeHost{err: sentinel})
	if !errors.Is(err, sentinel) || result.Status != catalogapi.StatusError || result.Error != "load failed" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	result, _ = Catalog(context.Background(), fakeHost{catalog: loaded})
	if got := Resolve(result, id[:10]); got.ConceptID != id || got.Status != catalogapi.StatusOK {
		t.Fatalf("resolution=%+v", got)
	}
	if got := Resolve(result, "c-00000000"); got.Status != catalogapi.StatusError {
		t.Fatalf("missing=%+v", got)
	}
}
