package execution

import (
	"encoding/json"
	"testing"

	"polyglot-rosetta/internal/domain/language"
)

func TestRequestJSONSerialization(t *testing.T) {
	req := Request{
		Language:   language.Go,
		SourceCode: "package main\nfunc Add(a, b int) int { return a + b }",
		IsSubmit:   true,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal Request: %v", err)
	}

	var decoded Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Request: %v", err)
	}

	if decoded.Language != req.Language || decoded.SourceCode != req.SourceCode || decoded.IsSubmit != req.IsSubmit {
		t.Errorf("Decoded request does not match original: %+v", decoded)
	}
}

func TestStatusJSONValues(t *testing.T) {
	for status, want := range map[Status]string{
		StatusPassed:  `"passed"`,
		StatusFailed:  `"failed"`,
		StatusError:   `"error"`,
		StatusTimeout: `"timeout"`,
	} {
		data, err := json.Marshal(status)
		if err != nil || string(data) != want {
			t.Errorf("Marshal(%q) = %q, %v; want %s", status, data, err, want)
		}
	}
}
