package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRequestsJSON(t *testing.T) {
	if RequestsJSON([]string{"--format=json"}) != true {
		t.Fatal("expected JSON for --format=json")
	}
	if RequestsJSON([]string{"--format", "json"}) != true {
		t.Fatal("expected JSON for --format json")
	}
	if RequestsJSON([]string{"--format=json", "--help"}) {
		t.Fatal("help should not request JSON")
	}
	if RequestsJSON([]string{"--format=text"}) {
		t.Fatal("text is not JSON")
	}
}

func TestValidateOutputFormat(t *testing.T) {
	if err := ValidateOutputFormat("json"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateOutputFormat("yaml"); err == nil || !strings.Contains(err.Error(), "unsupported output format") {
		t.Fatalf("error = %v", err)
	}
}

func TestReportJSON(t *testing.T) {
	var output bytes.Buffer
	err := ReportJSON(&output, map[string]string{"status": "error"}, errors.New("boom"))
	var reported *reportedError
	if !errors.As(err, &reported) {
		t.Fatalf("error = %v", err)
	}
	if !strings.Contains(output.String(), `"status":"error"`) {
		t.Fatalf("output = %q", output.String())
	}
}
