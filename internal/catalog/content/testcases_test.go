package content

import (
	"testing"
	"testing/fstest"
)

func TestLoadTestSuiteFormats(t *testing.T) {
	for _, test := range []struct {
		name    string
		content string
		concept string
		id      string
	}{
		{name: "array", content: `[{"id":1,"input":{},"expected":true}]`, concept: "hello-world", id: "1"},
		{name: "wrapped", content: `{"concept":"custom","test_cases":[{"id":"public-1"}]}`, concept: "custom", id: "public-1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			files := fstest.MapFS{"hello-world/testcases.json": &fstest.MapFile{Data: []byte(test.content)}}
			suite, err := LoadTestSuite(files, "hello-world")
			if err != nil {
				t.Fatal(err)
			}
			if suite.Concept != test.concept || len(suite.TestCases) != 1 || suite.TestCases[0].ID != test.id {
				t.Fatalf("suite = %+v", suite)
			}
		})
	}
}

func TestLoadTestSuiteRejectsInvalidCases(t *testing.T) {
	for _, test := range []struct {
		name    string
		content string
	}{
		{name: "malformed", content: `{"test_cases": [`},
		{name: "empty", content: `[]`},
		{name: "duplicate IDs", content: `[{"id":1},{"id":1}]`},
	} {
		t.Run(test.name, func(t *testing.T) {
			files := fstest.MapFS{"hello-world/testcases.json": &fstest.MapFile{Data: []byte(test.content)}}
			if _, err := LoadTestSuite(files, "hello-world"); err == nil {
				t.Fatal("expected test-case error")
			}
		})
	}
}
