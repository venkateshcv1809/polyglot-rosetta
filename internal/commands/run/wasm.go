//go:build js && wasm

package runcmd

import (
	"context"
	"fmt"
	"syscall/js"
	"time"

	executionapi "polyglot-rosetta/internal/api/execution"
	probeapi "polyglot-rosetta/internal/api/probe"
	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/domain/testcase"
	"polyglot-rosetta/internal/wasm"
)

// WASM only converts values and bridges host operations to the shared evaluator.
func WASM(store *wasm.CatalogStore) wasm.Binding {
	return wasm.Binding{Name: "run", Help: wasmHelp(), Handler: func(_ js.Value, args []js.Value) any {
		// Pin the entire host and catalog before yielding to JavaScript.
		snapshot, catalogErr := store.Catalog()
		ensure, ensureErr := wasm.EnsureAvailable()
		build, buildErr := wasm.BuildProgram()
		runCase, runErr := wasm.RunCase()
		dispose, disposeErr := wasm.DisposeProgram()
		var request runapi.Request
		var requestErr error
		if len(args) != 1 {
			requestErr = fmt.Errorf("run requires a request object")
		} else {
			requestErr = wasm.Decode(args[0], &request)
		}
		return wasm.Promise(func() any {
			failed := func(err error) runapi.Result { r := runapi.NewResult(); r.Error = err.Error(); return r }
			if requestErr != nil {
				return failed(requestErr)
			}
			for _, err := range []error{ensureErr, buildErr, runErr, disposeErr} {
				if err != nil {
					return failed(err)
				}
			}
			result, _ := Run(context.Background(), request, browserHost{
				catalog: snapshot, catalogErr: catalogErr, ensureAvailable: ensure,
				execution: browserExecutionHost{build: build, runCase: runCase, dispose: dispose},
			})
			return result
		})
	}}
}

type browserHost struct {
	catalog         *catalog.Catalog
	catalogErr      error
	ensureAvailable js.Value
	execution       browserExecutionHost
}

func (h browserHost) Resolve(_ context.Context, id string) (catalog.Concept, error) {
	if h.catalogErr != nil {
		return catalog.Concept{}, h.catalogErr
	}
	return h.catalog.ResolveConcept(id)
}

func (h browserHost) EnsureAvailable(_ context.Context, language string) error {
	value, err := wasm.InvokeWithTimeout(h.ensureAvailable, 60*time.Second, language)
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

func (h browserHost) Execute(ctx context.Context, request runapi.Request, tests []testcase.TestCase) (runapi.Result, error) {
	return executeEvaluation(ctx, request, tests, h.execution)
}

type browserExecutionHost struct{ build, runCase, dispose js.Value }

func (h browserExecutionHost) BuildProgram(ctx context.Context, program executionapi.Program) (executionapi.BuildResult, error) {
	var result executionapi.BuildResult
	if err := ctx.Err(); err != nil {
		return result, err
	}
	value, err := wasm.InvokeWithTimeout(h.build, time.Duration(program.TimeoutMs)*time.Millisecond, wasm.Value(program))
	if err != nil {
		return result, fmt.Errorf("buildProgram: %w", err)
	}
	if err := wasm.DecodeHostResult(value, &result, "BuildResult"); err != nil {
		return result, fmt.Errorf("buildProgram: %w", err)
	}
	return result, nil
}

func (h browserExecutionHost) RunCase(ctx context.Context, request executionapi.CaseRequest) (executionapi.CaseResult, error) {
	var result executionapi.CaseResult
	if err := ctx.Err(); err != nil {
		return result, err
	}
	// Allow a short response-delivery grace period after the host's case limit.
	value, err := wasm.InvokeWithTimeout(h.runCase, time.Duration(request.TimeoutMs)*time.Millisecond+time.Second, wasm.Value(request))
	if err != nil {
		return result, fmt.Errorf("runCase: %w", err)
	}
	if err := wasm.DecodeHostResult(value, &result, "CaseResult"); err != nil {
		return result, fmt.Errorf("runCase: %w", err)
	}
	return result, nil
}

func (h browserExecutionHost) DisposeProgram(_ context.Context, sessionID string) error {
	_, err := wasm.InvokeWithTimeout(h.dispose, 5*time.Second, sessionID)
	if err != nil {
		return fmt.Errorf("disposeProgram: %w", err)
	}
	return nil
}
