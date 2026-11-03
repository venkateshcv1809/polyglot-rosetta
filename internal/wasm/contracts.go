//go:build js && wasm

package wasm

// Runtime-discoverable fields for SDK authors. The Go wire types are defined in
// internal/api/execution; tests ensure these field names remain aligned.
var hostContracts = map[string][]Option{
	"Program": {
		{Name: "schema_version", Type: "integer", Required: true, Default: 1, Description: "Execution payload schema version."},
		{Name: "session_id", Type: "string", Required: true, Description: "Opaque per-run identifier allocated by Go; key all artifacts and resources by this ID."},
		{Name: "language", Type: "string", Required: true, Description: "Normalized solution language."},
		{Name: "entry", Type: "string", Required: true, Description: "Generated main file to compile or load."},
		{Name: "files", Type: "object<string,string>", Required: true, Description: "Generated solution, harness and supporting source files. Preserve their contents and entry-point behavior."},
		{Name: "build_required", Type: "boolean", Required: true, Description: "Whether this language has a separate timed compilation step. False for Python."},
		{Name: "timeout_ms", Type: "integer", Required: true, Description: "Maximum host build/preparation duration. Return failures as BuildResult.error with timing when measured."},
	},
	"BuildResult": {
		{Name: "schema_version", Type: "integer", Required: true, Default: 1, Description: "Echo the program schema version."},
		{Name: "session_id", Type: "string", Required: true, Description: "Echo the program session ID."},
		{Name: "build_duration_ms", Type: "integer", Description: "Nonnegative compiler duration; required for successful compiled programs and forbidden when build_required is false. Include for failed compiler attempts too."},
		{Name: "error", Type: "string", Description: "Build/preparation diagnostic; omit on success. Stops execution and triggers disposal."},
	},
	"CaseRequest": {
		{Name: "schema_version", Type: "integer", Required: true, Default: 1, Description: "Execution payload schema version."},
		{Name: "session_id", Type: "string", Required: true, Description: "The built session to execute."},
		{Name: "index", Type: "integer", Required: true, Description: "Zero-based position in the selected suite; not the authored testcase ID."},
		{Name: "args", Type: "string[]", Required: true, Description: "Entry-point arguments excluding argv[0]; currently the testcase index as a string."},
		{Name: "timeout_ms", Type: "integer", Required: true, Description: "Per-case deadline. Terminate the case and return timed_out: true when exceeded."},
		{Name: "max_output_bytes", Type: "integer", Required: true, Description: "Capture limit per stream; retain the tail and mark truncation. Do not count code-point characters as bytes."},
	},
	"CaseResult": {
		{Name: "schema_version", Type: "integer", Required: true, Default: 1, Description: "Echo the case schema version."},
		{Name: "session_id", Type: "string", Required: true, Description: "Echo the case session ID."},
		{Name: "index", Type: "integer", Required: true, Description: "Echo the case index."},
		{Name: "stdout", Type: "string", Required: true, Description: "Captured standard output including the unchanged harness result record."},
		{Name: "stderr", Type: "string", Required: true, Description: "Captured standard error; some harnesses report results on this stream."},
		{Name: "exit_code", Type: "integer", Required: true, Description: "Zero on normal exit, positive on failed exit, -1 if no exit code is available."},
		{Name: "timed_out", Type: "boolean", Required: true, Description: "Whether the case exceeded its deadline. Go records the case timeout and continues."},
		{Name: "stdout_truncated", Type: "boolean", Description: "Whether only the stdout tail was retained."},
		{Name: "stderr_truncated", Type: "boolean", Description: "Whether only the stderr tail was retained."},
		{Name: "error", Type: "string", Description: "Process/worker diagnostic. Reports a case crash, not a host callback failure."},
	},
}
