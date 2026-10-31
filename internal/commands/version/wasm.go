//go:build js && wasm

package versioncmd

import (
	"fmt"
	"syscall/js"

	"polyglot-rosetta/internal/wasm"
)

// WASM returns the JavaScript binding for the version operation.
func WASM() wasm.Binding {
	return wasm.Binding{Name: "version", Handler: wasmVersion, Help: wasmHelp()}
}

func wasmVersion(_ js.Value, args []js.Value) any {
	if len(args) != 0 {
		return wasm.Value(errorResult(fmt.Errorf("version accepts no arguments")))
	}
	return wasm.Value(Version())
}
