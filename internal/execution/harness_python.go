package execution

import (
	"encoding/json"
	"fmt"

	"polyglot-rosetta/internal/domain/testcase"
)

func preparePython(testCases []testcase.TestCase, marker string) (map[string]string, preparedProgram, error) {
	payload, err := json.Marshal(testCases)
	if err != nil {
		return nil, preparedProgram{}, fmt.Errorf("encode test cases: %w", err)
	}
	harness := fmt.Sprintf(`import json
import sys
import time
from solution import Solution

tests = json.loads(%s)
result_marker = %s
index = int(sys.argv[1])
test = tests[index]
solution = Solution()
started = time.perf_counter_ns()
actual = solution.solve(test["input"])
solution_time_ns = time.perf_counter_ns() - started
passed = actual == test["expected"]
print(f"\n{result_marker}:{index}:{str(passed).lower()}:{solution_time_ns}")
`, quoteJSON(string(payload)), quoteJSON(marker))

	return map[string]string{"harness.py": harness}, preparedProgram{
		run: Command{Name: "python3", Args: []string{"harness.py"}},
	}, nil
}
