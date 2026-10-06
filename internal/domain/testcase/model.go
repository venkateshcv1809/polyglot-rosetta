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

// GetTestCases returns either public-only or all test cases depending on the evaluation mode.
func (s *TestSuite) GetTestCases(includeHidden bool) []TestCase {
	if includeHidden {
		return s.TestCases
	}

	publicCases := make([]TestCase, 0)
	for _, tc := range s.TestCases {
		if !tc.Hidden {
			publicCases = append(publicCases, tc)
		}
	}
	return publicCases
}
