package execution

import (
	"encoding/json"
	"fmt"

	"polyglot-rosetta/internal/domain/testcase"
)

func prepareTypeScript(testCases []testcase.TestCase, marker string) (map[string]string, preparedProgram, error) {
	payload, err := json.Marshal(testCases)
	if err != nil {
		return nil, preparedProgram{}, fmt.Errorf("encode test cases: %w", err)
	}
	harness := fmt.Sprintf(`import { Solution } from "./solution";
declare const process: { argv: string[]; hrtime: { bigint(): bigint } };

const tests = JSON.parse(%s);
const resultMarker = %s;
const canonical = (value: unknown): string => {
  if (Array.isArray(value)) return "[" + value.map(canonical).join(",") + "]";
  if (value !== null && typeof value === "object") {
    const record = value as Record<string, unknown>;
    return "{" + Object.keys(record).sort().map((key) => JSON.stringify(key) + ":" + canonical(record[key])).join(",") + "}";
  }
  return JSON.stringify(value);
};

const index = Number(process.argv[2]);
const test: { input: unknown; expected: unknown } = tests[index];
const solution = new Solution();
const started = process.hrtime.bigint();
const actual = solution.solve(test.input);
const solutionTimeNs = process.hrtime.bigint() - started;
const passed = canonical(actual) === canonical(test.expected);
console.log("\n" + resultMarker + ":" + index + ":" + passed + ":" + solutionTimeNs);
`, quoteJSON(string(payload)), quoteJSON(marker))

	return map[string]string{"harness.ts": harness}, preparedProgram{
		build: []Command{{Name: "tsc", Args: []string{"--target", "ES2022", "--module", "commonjs", "--outDir", "build", "solution.ts", "harness.ts"}}},
		run:   Command{Name: "node", Args: []string{"build/harness.js"}},
	}, nil
}
