//go:build js && wasm

package runcmd

import (
	"context"
	"fmt"
	"syscall/js"

	probeapi "polyglot-rosetta/internal/api/probe"
	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/domain/testcase"
	"polyglot-rosetta/internal/wasm"
)

// WASM only converts values and bridges the browser host to the shared command.
func WASM(store *wasm.CatalogStore) wasm.Binding {
	return wasm.Binding{Name: "run", Help: wasmHelp(), Handler: func(_ js.Value, args []js.Value) any {
		// Pin the catalog and host callbacks before yielding to JavaScript.
		// Refreshing or rebinding must not change a run that is already in flight.
		snapshot, catalogErr := store.Catalog()
		ensureAvailable, ensureAvailableErr := wasm.EnsureAvailable()
		execute, executeErr := wasm.Execute()
		return wasm.Promise(func() any {
			failed := func(err error) runapi.Result { r := runapi.NewResult(); r.Error = err.Error(); return r }
			if len(args) != 1 {
				return failed(fmt.Errorf("run requires a request object"))
			}
			var request runapi.Request
			if err := wasm.Decode(args[0], &request); err != nil {
				return failed(err)
			}
			// Require a complete host before running a command.
			if ensureAvailableErr != nil {
				return failed(ensureAvailableErr)
			}
			if executeErr != nil {
				return failed(executeErr)
			}
			result, _ := Run(context.Background(), request, browserHost{
				catalog: snapshot, catalogErr: catalogErr, ensureAvailable: ensureAvailable, execute: execute,
			})
			return result
		})
	}}
}

type browserHost struct {
	catalog         *catalog.Catalog
	catalogErr      error
	ensureAvailable js.Value
	execute         js.Value
}

func (h browserHost) Resolve(_ context.Context, id string) (catalog.Concept, error) {
	if h.catalogErr != nil {
		return catalog.Concept{}, h.catalogErr
	}
	return h.catalog.ResolveConcept(id)
}

func (h browserHost) EnsureAvailable(_ context.Context, language string) error {
	value, err := wasm.Invoke(h.ensureAvailable, language)
	if err != nil {
		return err
	}
	var snapshot probeapi.CacheSnapshot
	if err := wasm.Decode(value, &snapshot); err != nil {
		return err
	}
	if !snapshot.Ready(language) {
		return fmt.Errorf("language %q is not ready", language)
	}
	return nil
}

func (h browserHost) Execute(_ context.Context, request runapi.Request, tests []testcase.TestCase) (runapi.Result, error) {
	value, err := wasm.Invoke(h.execute, wasm.Value(request), wasm.Value(tests))
	if err != nil {
		return runapi.Result{}, err
	}
	var result runapi.Result
	err = wasm.Decode(value, &result)
	return result, err
}
