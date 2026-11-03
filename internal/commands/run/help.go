//go:build js && wasm

package runcmd

import (
	"polyglot-rosetta/internal/domain/evaluation"
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
			{Name: "mode", Type: "string", Default: string(evaluation.ModePublic), Values: []string{string(evaluation.ModePublic), string(evaluation.ModeAll)}, Description: "public selects visible cases; all also includes hidden cases."},
		}}},
		Callbacks: []string{"ensureAvailable", "buildProgram", "runCase", "disposeProgram"},
		Returns:   "Promise<RunResult> {schema_version, id, language, mode, status, total_passed, total_failed, duration_ms, build_duration_ms?, test_results, error?}",
		Notes:     []string{"Go generates sources, selects and orders cases, parses outputs, compares answers and aggregates results. Hosts only build, run a single case and dispose sessions.", "Activate the requested category before running. Host deadlines must terminate or isolate work; the core's Promise deadlines cannot interrupt synchronous JavaScript.", "Each run gets a unique session_id and pins callbacks and catalog before asynchronous work. disposeProgram is called after any attempted build, including rejected or invalid build responses.", "Student crashes and timeouts are raw CaseResults and do not stop subsequent cases. Rejected callbacks or invalid responses abort the run and trigger disposal.", "Command failures resolve to status: error. Unexpected bridge failures reject the promise."},
	}
}
