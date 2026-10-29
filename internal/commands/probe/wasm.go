//go:build js && wasm

package probecmd

import (
	"fmt"
	"syscall/js"

	probeapi "polyglot-rosetta/internal/api/probe"
	"polyglot-rosetta/internal/wasm"
)

type refreshPayload struct {
	Toolchains []probeapi.Toolchain `json:"toolchains"`
}

// refreshHost adapts toolchains already returned by the browser refresh callback.
type refreshHost struct {
	rows []probeapi.Toolchain
}

func (host refreshHost) RefreshCache() ([]probeapi.Toolchain, error) {
	return host.rows, nil
}

// WASM returns a binding that probes via the host refreshCache callback.
func WASM() wasm.Binding {
	return wasm.Binding{Name: "probe", Help: wasmHelp(), Handler: func(_ js.Value, args []js.Value) any {
		return wasm.Promise(func() any { return wasmProbe(args) })
	}}
}

func wasmProbe(args []js.Value) probeapi.Result {
	if len(args) > 1 {
		return errorResult(fmt.Errorf("probe accepts an optional language or request object"))
	}
	var request probeapi.Request
	if len(args) == 1 {
		if args[0].Type() == js.TypeString {
			request.Language = args[0].String()
		} else if args[0].Type() == js.TypeObject {
			if err := wasm.Decode(args[0], &request); err != nil {
				return errorResult(err)
			}
		} else {
			return errorResult(fmt.Errorf("probe language must be a string or request object"))
		}
	}
	if request.Clear {
		if request.Language != "" {
			return errorResult(fmt.Errorf("clear cannot be combined with language: the cache covers all languages"))
		}
		callback, err := wasm.ClearCache()
		if err != nil {
			return errorResult(err)
		}
		return ClearCache(clearHost{callback: callback})
	}
	refreshCache, err := wasm.RefreshCache()
	if err != nil {
		return errorResult(err)
	}
	value, err := wasm.Invoke(refreshCache)
	if err != nil {
		return errorResult(err)
	}
	var payload refreshPayload
	if err := wasm.Decode(value, &payload); err != nil {
		return errorResult(err)
	}
	if payload.Toolchains == nil {
		payload.Toolchains = []probeapi.Toolchain{}
	}
	return Probe(request.Language, refreshHost{rows: payload.Toolchains})
}

type clearHost struct{ callback js.Value }

func (host clearHost) ClearCache() error {
	_, err := wasm.Invoke(host.callback)
	return err
}
