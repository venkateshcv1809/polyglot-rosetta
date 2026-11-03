// Package execution defines the host operations used by the shared evaluator.
// Hosts compile and launch programs; they never select cases or grade answers.
package execution

// Program contains generated sources for one isolated evaluation session.
// Files include the solution and harness, but never the expected answers.
type Program struct {
	SchemaVersion int               `json:"schema_version"`
	SessionID     string            `json:"session_id"`
	Language      string            `json:"language"`
	Entry         string            `json:"entry"`
	Files         map[string]string `json:"files"`
	BuildRequired bool              `json:"build_required"`
	TimeoutMs     int64             `json:"timeout_ms"`
}

// BuildResult reports preparation or compilation of a host session. Hosts must
// release even partially prepared sessions when disposeProgram is called.
type BuildResult struct {
	SchemaVersion   int    `json:"schema_version"`
	SessionID       string `json:"session_id"`
	BuildDurationMs *int64 `json:"build_duration_ms,omitempty"`
	Error           string `json:"error,omitempty"`
}

// CaseRequest launches the built entry point in a fresh execution environment.
// Args excludes argv[0]; the sole argument is the zero-based selected-case index.
type CaseRequest struct {
	SchemaVersion  int      `json:"schema_version"`
	SessionID      string   `json:"session_id"`
	Index          int      `json:"index"`
	Args           []string `json:"args"`
	TimeoutMs      int64    `json:"timeout_ms"`
	MaxOutputBytes int      `json:"max_output_bytes"`
}

// CaseResult is raw process/worker output, not a scored evaluation result.
// Student crashes and deadlines are returned here; callback failures indicate
// host infrastructure errors and abort the session.
type CaseResult struct {
	SchemaVersion   int    `json:"schema_version"`
	SessionID       string `json:"session_id"`
	Index           int    `json:"index"`
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	ExitCode        int    `json:"exit_code"`
	TimedOut        bool   `json:"timed_out"`
	StdoutTruncated bool   `json:"stdout_truncated,omitempty"`
	StderrTruncated bool   `json:"stderr_truncated,omitempty"`
	Error           string `json:"error,omitempty"`
}
