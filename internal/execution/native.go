package execution

import (
	"context"
	"fmt"
	"time"

	"polyglot-rosetta/internal/api"
	executionapi "polyglot-rosetta/internal/api/execution"
	"polyglot-rosetta/internal/domain/language"
)

// A native host belongs to one evaluation; parallel runs never share a workspace.
type nativeHost struct {
	workspace *Workspace
	program   preparedProgram
}

// NewNativeHost returns a fresh session host for an evaluation on this machine.
func NewNativeHost() Host { return &nativeHost{} }

func (h *nativeHost) BuildProgram(ctx context.Context, program executionapi.Program) (executionapi.BuildResult, error) {
	result := executionapi.BuildResult{SchemaVersion: api.SchemaVersion, SessionID: program.SessionID}
	workspace, err := NewWorkspace()
	if err != nil {
		return result, err
	}
	h.workspace = workspace
	for name, content := range program.Files {
		if _, err := workspace.WriteFile(name, content); err != nil {
			return result, err
		}
	}
	definition, ok := languagePrograms[language.Language(program.Language)]
	if !ok {
		return result, fmt.Errorf("unsupported native language %q", program.Language)
	}
	h.program = preparedProgram{run: definition.run}
	h.program.run.Dir = workspace.Dir
	if definition.build != nil {
		command := *definition.build
		command.Dir = workspace.Dir
		h.program.build = &command
	}
	duration, err := buildProgram(ctx, h.program.build, time.Duration(program.TimeoutMs)*time.Millisecond)
	if duration != nil {
		ms := duration.Milliseconds()
		result.BuildDurationMs = &ms
	}
	if err != nil {
		result.Error = err.Error()
	}
	return result, nil
}

func (h *nativeHost) RunCase(ctx context.Context, request executionapi.CaseRequest) (executionapi.CaseResult, error) {
	command := h.program.run
	command.Args = append(append([]string(nil), command.Args...), request.Args...)
	command.MaxOutputBytes = request.MaxOutputBytes
	process := RunCommand(ctx, time.Duration(request.TimeoutMs)*time.Millisecond, command)
	result := executionapi.CaseResult{
		SchemaVersion: api.SchemaVersion, SessionID: request.SessionID, Index: request.Index,
		Stdout: process.Stdout, Stderr: process.Stderr, ExitCode: process.ExitCode, TimedOut: process.TimedOut,
		StdoutTruncated: process.StdoutTruncated, StderrTruncated: process.StderrTruncated,
	}
	if process.Error != nil {
		result.Error = process.Error.Error()
	}
	return result, nil
}

func (h *nativeHost) DisposeProgram(_ context.Context, _ string) error {
	if h.workspace == nil {
		return nil
	}
	return h.workspace.Cleanup()
}
