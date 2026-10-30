package cli

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestPromptForm(t *testing.T) {
	fields := []PromptField{
		{Key: "id", Prompt: "ID", Required: true},
		{Key: "title", Prompt: "Title", DefaultValue: "Untitled"},
	}
	var output bytes.Buffer

	result, err := PromptForm(strings.NewReader("\narrays\n\n"), &output, fields, nil)
	if err != nil {
		t.Fatalf("PromptForm: %v", err)
	}
	if result["id"] != "arrays" || result["title"] != "Untitled" {
		t.Fatalf("result = %#v", result)
	}
	wantOutput := "ID:   Error: this field is required.\nID: Title [Untitled]: "
	if output.String() != wantOutput {
		t.Fatalf("output:\n%q\nwant:\n%q", output.String(), wantOutput)
	}
}

func TestPromptFormPreservesExistingValues(t *testing.T) {
	var output bytes.Buffer
	result, err := PromptForm(strings.NewReader(""), &output, []PromptField{{Key: "id", Prompt: "ID", Required: true}}, map[string]string{"id": "arrays"})
	if err != nil {
		t.Fatalf("PromptForm: %v", err)
	}
	if result["id"] != "arrays" || output.Len() != 0 {
		t.Fatalf("result = %#v, output = %q", result, output.String())
	}
}

func TestPromptFormReturnsEOF(t *testing.T) {
	_, err := PromptForm(strings.NewReader(""), io.Discard, []PromptField{{Key: "id", Prompt: "ID", Required: true}}, nil)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("error = %v, want io.EOF", err)
	}
}

func TestPromptFormReturnsWriteError(t *testing.T) {
	want := errors.New("write failed")
	_, err := PromptForm(strings.NewReader("value\n"), errorWriter{err: want}, []PromptField{{Key: "id", Prompt: "ID"}}, nil)
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

func TestPromptFormReturnsReadError(t *testing.T) {
	want := errors.New("read failed")
	_, err := PromptForm(errorReader{err: want}, io.Discard, []PromptField{{Key: "id", Prompt: "ID"}}, nil)
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}

type errorReader struct {
	err error
}

func (r errorReader) Read([]byte) (int, error) {
	return 0, r.err
}
