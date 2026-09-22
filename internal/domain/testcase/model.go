package testcase

type TestCase struct {
	ID          string         `json:"id"`
	Description string         `json:"description"`
	Input       map[string]any `json:"input"`
	Expected    any            `json:"expected"`
	Hidden      bool           `json:"hidden"`
}

type TestSuite struct {
	Concept   string     `json:"concept"`
	TestCases []TestCase `json:"test_cases"`
}
