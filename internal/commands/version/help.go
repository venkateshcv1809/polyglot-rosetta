//go:build js && wasm

package versioncmd

import (
	"polyglot-rosetta/internal/wasm"
)

// wasmHelp describes the public WASM operation separately from its handler.
func wasmHelp() wasm.MethodHelp {
	return wasm.MethodHelp{
		Description: "Return the embedded version of the running Go core.",
		Signatures:  []string{"version()"},
		Returns:     "VersionResult {schema_version, status, version?, error?}",
	}
}
