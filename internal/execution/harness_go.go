package execution

import (
	"fmt"
	"strings"

	"polyglot-rosetta/internal/domain/testcase"
)

func prepareGo(testCases []testcase.TestCase, marker string) (map[string]string, preparedProgram, error) {
	var harness strings.Builder
	harness.WriteString(`package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"
)

const resultMarker = "`)
	harness.WriteString(marker)
	harness.WriteString(`"

func normalize(value any) any {
	data, err := json.Marshal(value)
	if err != nil { return value }
	var normalized any
	if err := json.Unmarshal(data, &normalized); err != nil { return value }
	return normalized
}

func runCase(index int, inputJSON, expectedJSON string) {
	var input any
	var expected any
	if json.Unmarshal([]byte(inputJSON), &input) != nil { return }
	if json.Unmarshal([]byte(expectedJSON), &expected) != nil { return }
	started := time.Now()
	actual := Solution{}.Solve(input)
	solutionTimeNs := time.Since(started).Nanoseconds()
	passed := reflect.DeepEqual(normalize(actual), expected)
	fmt.Printf("\n%s:%d:%t:%d\n", resultMarker, index, passed, solutionTimeNs)
}

func main() {
	if len(os.Args) != 2 { return }
	switch os.Args[1] {
`)
	for index, testCase := range testCases {
		input, expected, err := encodeCase(testCase)
		if err != nil {
			return nil, preparedProgram{}, err
		}
		fmt.Fprintf(&harness, "\tcase %q: runCase(%d, %q, %q)\n", fmt.Sprint(index), index, input, expected)
	}
	harness.WriteString("\t}\n}\n")

	files := map[string]string{
		"go.mod":  "module prosetta-evaluation\n\ngo 1.22\n",
		"main.go": harness.String(),
	}
	return files, preparedProgram{
		build: []Command{{Name: "go", Args: []string{"build", "-buildvcs=false", "-o", "solution", "."}}},
		run:   Command{Name: "./solution"},
	}, nil
}
