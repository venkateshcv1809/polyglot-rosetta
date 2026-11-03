package execution

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// Harnesses report a single line: marker:index:nanoseconds:format:hex-output.
// Hex keeps multiline strings, Unicode, and arbitrary debug output separate
// from the result framing. The marker prevents accidental log collisions.
type caseResult struct {
	Output         []byte
	Format         string
	SolutionTimeNs int64
}

func newResultMarker() (string, error) {
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", fmt.Errorf("create evaluation result marker: %w", err)
	}
	return "__PROSETTA_RESULT_" + hex.EncodeToString(token[:]) + "__", nil
}

func parseCaseOutcome(index int, output, marker string) (caseResult, error) {
	var outcome caseResult
	found := false
	for _, outputLine := range strings.Split(output, "\n") {
		line := strings.TrimSpace(outputLine)
		if !strings.HasPrefix(line, marker+":") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 5 {
			return outcome, fmt.Errorf("invalid evaluation result %q", line)
		}
		reportedIndex, err := strconv.Atoi(parts[1])
		if err != nil || reportedIndex != index {
			return outcome, fmt.Errorf("invalid testcase index in evaluation result %q", line)
		}
		duration, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || duration < 0 {
			return outcome, fmt.Errorf("invalid solution time in evaluation result %q", line)
		}
		if parts[3] != "json" && parts[3] != "text" {
			return outcome, fmt.Errorf("invalid output format in evaluation result %q", line)
		}
		actual, err := hex.DecodeString(parts[4])
		if err != nil {
			return outcome, fmt.Errorf("invalid output encoding in evaluation result %q", line)
		}
		if found {
			return outcome, fmt.Errorf("duplicate result for testcase index %d", index)
		}
		found = true
		outcome.Output, outcome.Format = actual, parts[3]
		outcome.SolutionTimeNs = duration
	}
	if !found {
		return outcome, fmt.Errorf("execution did not report testcase index %d", index)
	}
	return outcome, nil
}
