package execution

import (
	"testing"
)

func TestParseCaseOutcomeRejectsMissingAndDuplicateResults(t *testing.T) {
	marker := "test-marker"
	for _, output := range []string{
		"no result",
		marker + ":0:123:json:74727565\n" + marker + ":0:456:json:66616c7365\n",
	} {
		if _, err := parseCaseOutcome(0, output, marker); err == nil {
			t.Fatalf("parseCaseOutcome(%q) returned nil error", output)
		}
	}
}

func TestParseCaseOutcomeIgnoresForeignMarkers(t *testing.T) {
	marker := "random-marker"
	output := "__PROSETTA_RESULT__:0:9:json:66616c7365\n" + marker + ":0:123:json:74727565\n"

	outcome, err := parseCaseOutcome(0, output, marker)
	if err != nil {
		t.Fatalf("parseCaseOutcome: %v", err)
	}
	if string(outcome.Output) != "true" || outcome.Format != "json" || outcome.SolutionTimeNs != 123 {
		t.Fatalf("outcome = %+v", outcome)
	}
}

func TestParseCaseOutcomeRejectsMalformedRecords(t *testing.T) {
	marker := "test-marker"
	for _, output := range []string{
		marker + ":x:1:json:74727565",
		marker + ":1:1:json:74727565",
		marker + ":0:extra:json:74727565",
		marker + ":0:-1:json:74727565",
		marker + ":0:9223372036854775808:json:74727565",
		marker + ":0:1:invalid:74727565",
		marker + ":0:1:json:xyz",
		marker + ":0:1:json:1",
		marker + ":0:1:json",
	} {
		if _, err := parseCaseOutcome(0, output, marker); err == nil {
			t.Errorf("parseCaseOutcome(%q) returned nil error", output)
		}
	}
}
