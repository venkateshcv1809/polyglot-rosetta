package execution

import (
	"testing"
)

func TestParseCaseOutcomeRejectsMissingAndDuplicateResults(t *testing.T) {
	marker := "test-marker"
	for _, output := range []string{
		"no result",
		marker + ":0:true:123\n" + marker + ":0:false:456\n",
	} {
		if _, err := parseCaseOutcome(0, output, marker); err == nil {
			t.Fatalf("parseCaseOutcome(%q) returned nil error", output)
		}
	}
}

func TestParseCaseOutcomeIgnoresForeignMarkers(t *testing.T) {
	marker := "random-marker"
	output := "__PROSETTA_RESULT__:0:false:9\n" + marker + ":0:true:123\n"

	outcome, err := parseCaseOutcome(0, output, marker)
	if err != nil {
		t.Fatalf("parseCaseOutcome: %v", err)
	}
	if !outcome.Passed || outcome.SolutionTimeNs != 123 {
		t.Fatalf("outcome = %+v", outcome)
	}
}

func TestParseCaseOutcomeRejectsMalformedRecords(t *testing.T) {
	marker := "test-marker"
	for _, output := range []string{
		marker + ":x:true:1",
		marker + ":1:true:1",
		marker + ":0:maybe:1",
		marker + ":0:true:extra",
		marker + ":0:true:-1",
		marker + ":0:true:9223372036854775808",
		marker + ":0:true",
	} {
		if _, err := parseCaseOutcome(0, output, marker); err == nil {
			t.Errorf("parseCaseOutcome(%q) returned nil error", output)
		}
	}
}
