//go:build js && wasm

// Command wasm exposes the execution command to JavaScript hosts.
package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/catalog"
	catalogbundle "polyglot-rosetta/internal/catalog/bundle"
	"polyglot-rosetta/internal/version"
)

var registeredFunctions []js.Func
var loadedCatalog *catalog.Catalog

func main() {
	core := js.Global().Get("Object").New()
	registerFunction(core, "version", wasmVersion)
	registerFunction(core, "validateRequest", wasmValidateRequest)
	registerFunction(core, "loadCatalog", wasmLoadCatalog)
	registerFunction(core, "catalog", func(js.Value, []js.Value) any {
		if loadedCatalog == nil {
			return jsValueResult(nil, fmt.Errorf("catalog content is not loaded"))
		}
		return jsValueResult(loadedCatalog, nil)
	})
	registerFunction(core, "testCases", func(_ js.Value, args []js.Value) any {
		if loadedCatalog == nil {
			return jsValueResult(nil, fmt.Errorf("catalog content is not loaded"))
		}
		if len(args) != 2 {
			return jsValueResult(nil, fmt.Errorf("testCases requires a concept ID and evaluation mode"))
		}
		concept, err := loadedCatalog.ResolveConcept(args[0].String())
		if err != nil {
			return jsValueResult(nil, err)
		}
		mode := runapi.Mode(args[1].String())
		if mode != runapi.ModePublic && mode != runapi.ModeAll {
			return jsValueResult(nil, fmt.Errorf("invalid evaluation mode %q", mode))
		}
		tests := make([]any, 0, len(concept.TestCases))
		for _, test := range concept.TestCases {
			if mode == runapi.ModeAll || !test.Hidden {
				tests = append(tests, test)
			}
		}
		return jsValueResult(tests, nil)
	})
	js.Global().Set("ProsettaWASM", core)
	select {}
}

func wasmLoadCatalog(_ js.Value, args []js.Value) any {
	if len(args) != 3 || args[2].Type() != js.TypeObject {
		return jsResult(false, fmt.Errorf("loadCatalog requires index, templates, and category archives"))
	}
	index, err := copyJSBytes(args[0])
	if err != nil {
		return jsResult(false, fmt.Errorf("read index: %w", err))
	}
	templates, err := copyJSBytes(args[1])
	if err != nil {
		return jsResult(false, fmt.Errorf("read templates: %w", err))
	}
	keys := js.Global().Get("Object").Call("keys", args[2])
	archives := make(map[string][]byte, keys.Length())
	for index := 0; index < keys.Length(); index++ {
		name := keys.Index(index).String()
		data, err := copyJSBytes(args[2].Get(name))
		if err != nil {
			return jsResult(false, fmt.Errorf("read archive %q: %w", name, err))
		}
		archives[name] = data
	}
	bundle, err := catalogbundle.Load(index, templates, archives)
	if err != nil {
		return jsResult(false, err)
	}
	loadedCatalog = &bundle.Catalog
	return jsResult(true, nil)
}

func copyJSBytes(value js.Value) ([]byte, error) {
	if !value.InstanceOf(js.Global().Get("Uint8Array")) || value.Length() == 0 || value.Length() > 128<<20 {
		return nil, fmt.Errorf("expected a non-empty Uint8Array within the size limit")
	}
	data := make([]byte, value.Length())
	if copied := js.CopyBytesToGo(data, value); copied != len(data) {
		return nil, fmt.Errorf("incomplete byte copy")
	}
	return data, nil
}

func registerFunction(target js.Value, name string, handler func(js.Value, []js.Value) any) {
	function := js.FuncOf(handler)
	registeredFunctions = append(registeredFunctions, function)
	target.Set(name, function)
}

func wasmVersion(_ js.Value, _ []js.Value) any {
	return version.Value
}

func jsValueResult(value any, err error) js.Value {
	if err != nil {
		return jsResult(false, err)
	}
	data, marshalErr := json.Marshal(value)
	if marshalErr != nil {
		return jsResult(false, marshalErr)
	}
	result := jsResult(true, nil)
	result.Set("value", js.Global().Get("JSON").Call("parse", string(data)))
	return result
}

func wasmValidateRequest(_ js.Value, args []js.Value) any {
	if len(args) != 1 || args[0].Type() != js.TypeObject {
		return jsResult(false, fmt.Errorf("validateRequest requires one request object"))
	}

	encoded := js.Global().Get("JSON").Call("stringify", args[0])
	if encoded.Type() != js.TypeString {
		return jsResult(false, fmt.Errorf("request must be JSON serializable"))
	}

	var request runapi.Request
	if err := json.Unmarshal([]byte(encoded.String()), &request); err != nil {
		return jsResult(false, fmt.Errorf("decode request: %w", err))
	}
	if err := request.Validate(); err != nil {
		return jsResult(false, err)
	}
	return jsResult(true, nil)
}

func jsResult(ok bool, err error) js.Value {
	result := js.Global().Get("Object").New()
	result.Set("ok", ok)
	if err != nil {
		result.Set("error", err.Error())
	}
	return result
}
