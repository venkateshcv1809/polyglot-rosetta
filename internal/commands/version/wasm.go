//go:build js && wasm

package versioncmd

import (
	"syscall/js"

	"polyglot-rosetta/internal/wasm"
)

// WASM returns the JavaScript binding for the version operation.
func WASM() wasm.Binding {
	return wasm.Binding{Name: "version", Handler: wasmVersion}
}

func wasmVersion(_ js.Value, _ []js.Value) any {
	return versionResult().Version
}
