package cli

import (
	"errors"
	"testing"
)

func TestHandle(t *testing.T) {
	type options struct {
		value string
	}

	var got options
	handler := Handle(
		func(commandPath string, args []string) (options, error) {
			if commandPath != "app command" {
				t.Fatalf("command path = %q", commandPath)
			}
			return options{value: args[0]}, nil
		},
		func(opts options) error {
			got = opts
			return nil
		},
	)

	if err := handler("app command", []string{"value"}); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.value != "value" {
		t.Fatalf("options = %+v", got)
	}
}

func TestHandleReturnsParseErrorWithoutRunning(t *testing.T) {
	want := errors.New("parse failed")
	runCalled := false
	handler := Handle(
		func(string, []string) (struct{}, error) { return struct{}{}, want },
		func(struct{}) error {
			runCalled = true
			return nil
		},
	)

	if err := handler("command", nil); !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
	if runCalled {
		t.Fatal("run called after parse failure")
	}
}

func TestHandleReturnsRunError(t *testing.T) {
	want := errors.New("run failed")
	handler := Handle(
		func(string, []string) (struct{}, error) { return struct{}{}, nil },
		func(struct{}) error { return want },
	)

	if err := handler("command", nil); !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}
