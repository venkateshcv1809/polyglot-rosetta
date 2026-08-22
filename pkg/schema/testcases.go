package schema

// TestCase models a single execution input/output assertion.
type TestCase struct {
	ID       int    `json:"id"`
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

// TestCasesFile wraps the array for testcases.json.
type TestCasesFile []TestCase
