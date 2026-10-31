package catalogcmd

import (
	"context"
	"fmt"

	catalogapi "polyglot-rosetta/internal/api/catalog"
	"polyglot-rosetta/internal/catalog"
)

// Host supplies validated catalog content using its environment's storage.
type Host interface {
	Activate(context.Context) (string, *catalog.Catalog, error)
}

// Catalog activates a host snapshot and constructs the shared command result.
func Catalog(ctx context.Context, host Host) (catalogapi.Result, error) {
	result := catalogapi.NewResult(catalogapi.StatusError)
	root, loaded, err := host.Activate(ctx)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}
	result.Status, result.Root, result.Catalog = catalogapi.StatusOK, root, loaded
	return result, nil
}

// Resolve selects a concept using the same full-ID/prefix rules as run.
func Resolve(result catalogapi.Result, id string) catalogapi.Result {
	if result.Status != catalogapi.StatusOK {
		return result
	}
	var err error
	if result.Catalog == nil {
		err = fmt.Errorf("catalog metadata is not loaded")
	} else {
		var concept catalog.Concept
		concept, err = result.Catalog.ResolveConcept(id)
		if err == nil {
			result.ConceptID, err = catalog.ConceptID(concept.Path)
		}
	}
	if err != nil {
		failed := catalogapi.NewResult(catalogapi.StatusError)
		failed.Error = err.Error()
		return failed
	}
	return result
}
