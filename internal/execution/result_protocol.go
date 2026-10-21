package execution

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	domainexecution "polyglot-rosetta/internal/domain/execution"
)

func newResultMarker() (string, error) {
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", fmt.Errorf("create evaluation result marker: %w", err)
	}
	return "__PROSETTA_RESULT_" + hex.EncodeToString(token[:]) + "__", nil
}

func parseCaseOutcome(index int, output, marker string) (domainexecution.TestOutcome, error) {
	var outcome domainexecution.TestOutcome
	found := false
	for _, outputLine := range strings.Split(output, "\n") {
		line := strings.TrimSpace(outputLine)
		if !strings.HasPrefix(line, marker+":") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 4 {
			return outcome, fmt.Errorf("invalid evaluation result %q", line)
		}
		reportedIndex, err := strconv.Atoi(parts[1])
		if err != nil || reportedIndex != index {
			return outcome, fmt.Errorf("invalid testcase index in evaluation result %q", line)
		}
		passed, err := strconv.ParseBool(parts[2])
		if err != nil {
			return outcome, fmt.Errorf("invalid status in evaluation result %q", line)
		}
		duration, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil || duration < 0 {
			return outcome, fmt.Errorf("invalid solution time in evaluation result %q", line)
		}
		if found {
			return outcome, fmt.Errorf("duplicate result for testcase index %d", index)
		}
		found = true
		outcome.Passed = passed
		outcome.SolutionTimeNs = duration
	}
	if !found {
		return outcome, fmt.Errorf("execution did not report testcase index %d", index)
	}
	return outcome, nil
}
