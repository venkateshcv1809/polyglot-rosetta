package cli

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"text/template"
)

//go:embed output/failure.tmpl
var failureTemplateFS embed.FS

type commandFailure struct {
	action string
	cause  error
}

type reportedError struct{ cause error }

func (e *reportedError) Error() string { return e.cause.Error() }
func (e *reportedError) Unwrap() error { return e.cause }

func (f *commandFailure) Error() string { return fmt.Sprintf("%s: %v", f.action, f.cause) }
func (f *commandFailure) Unwrap() error { return f.cause }

// CommandFailure adds a user-facing action to an error without losing its cause.
// ExitCode renders the action and cause on separate lines.
func CommandFailure(action string, err error) error {
	if err == nil {
		return nil
	}
	return &commandFailure{action: action, cause: err}
}

// ReportedError preserves a failure's exit status without printing it again.
// Use it when a command has already written a machine-readable error result.
func ReportedError(err error) error {
	if err == nil {
		return nil
	}
	return &reportedError{cause: err}
}

func printFailure(output io.Writer, err error) {
	data := struct {
		Action string
		Reason string
	}{Action: err.Error()}
	var failure *commandFailure
	if errors.As(err, &failure) {
		data.Action = failure.action
		data.Reason = failure.cause.Error()
	}

	tmpl, parseErr := template.New("failure.tmpl").ParseFS(failureTemplateFS, "output/failure.tmpl")
	if parseErr != nil {
		fmt.Fprintf(output, "ERROR  %v\n", err)
		return
	}
	if renderErr := tmpl.Execute(output, data); renderErr != nil {
		fmt.Fprintf(output, "ERROR  %v\n", err)
	}
}
