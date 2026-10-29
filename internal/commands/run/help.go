//go:build js && wasm

package runcmd

import (
	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/wasm"
)

// wasmHelp describes the public WASM operation separately from its handler.
func wasmHelp() wasm.MethodHelp {
	return wasm.MethodHelp{
		Description: "Evaluate submitted source against selected cases in an activated catalog.",
		Async:       true,
		Signatures:  []string{"run({id, language, source, mode?})"},
		Parameters: []wasm.Option{{Name: "request", Type: "object", Required: true, Description: "Solution evaluation request.", Properties: []wasm.Option{
			{Name: "id", Type: "string", Required: true, Description: "Full concept ID or unique prefix of at least eight hexadecimal digits."},
			{Name: "language", Type: "string", Required: true, Description: "Supported language; surrounding whitespace and case are normalized."},
			{Name: "source", Type: "string", Required: true, Description: "Non-empty solution source code, not a file path."},
			{Name: "mode", Type: "string", Default: string(runapi.ModePublic), Values: []string{string(runapi.ModePublic), string(runapi.ModeAll)}, Description: "public selects visible cases; all also includes hidden cases."},
		}}},
		Callbacks: []string{"ensureAvailable", "execute"},
		Returns:   "Promise<RunResult> {schema_version, id, language, mode, status, total_passed, total_failed, duration_ms, test_results, error?}",
		Notes:     []string{"Activate the requested category with catalog before running. The host owns compilation, workers, and deadlines.", "Command failures resolve to status: error. Unexpected bridge failures reject the promise."},
	}
}
