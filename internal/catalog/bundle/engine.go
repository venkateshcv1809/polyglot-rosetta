package bundle

import (
	"fmt"
	"strconv"
	"strings"
)

// CompatibleEngine reports whether engine can load a catalog that requires minEngine.
// An empty minEngine accepts any engine. Engine "dev" always passes.
func CompatibleEngine(engine, minEngine string) error {
	required := strings.TrimSpace(strings.TrimPrefix(minEngine, "v"))
	if required == "" {
		return nil
	}
	current := strings.TrimSpace(strings.TrimPrefix(engine, "v"))
	if current == "" || current == "dev" {
		return nil
	}
	if compareVersions(current, required) < 0 {
		return fmt.Errorf("catalog requires Prosetta %s or newer (running %s)", minEngine, engine)
	}
	return nil
}

func compareVersions(left, right string) int {
	leftParts := versionParts(left)
	rightParts := versionParts(right)
	n := len(leftParts)
	if len(rightParts) > n {
		n = len(rightParts)
	}
	for i := 0; i < n; i++ {
		var a, b int
		if i < len(leftParts) {
			a = leftParts[i]
		}
		if i < len(rightParts) {
			b = rightParts[i]
		}
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
	}
	return 0
}

func versionParts(value string) []int {
	value = strings.SplitN(value, "-", 2)[0]
	value = strings.SplitN(value, "+", 2)[0]
	fields := strings.Split(value, ".")
	parts := make([]int, 0, len(fields))
	for _, field := range fields {
		n, err := strconv.Atoi(field)
		if err != nil {
			parts = append(parts, 0)
			continue
		}
		parts = append(parts, n)
	}
	return parts
}
