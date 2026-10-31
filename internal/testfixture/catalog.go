// Package testfixture supplies isolated inputs shared by integration tests.
package testfixture

import (
	"os"
	"path/filepath"
	"testing"
)

// Catalog creates a small authoring tree and starter templates owned by the
// calling test. Tests must not depend on the repository's live problems tree.
func Catalog(t testing.TB) (problems, templates string) {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"problems/info.json":                           `{"id":"root","title":"Test catalog","items":[{"id":"fundamentals","order":1}]}`,
		"problems/fundamentals/info.json":              `{"id":"fundamentals","title":"Fundamentals","items":[{"id":"hello-world","order":1}]}`,
		"problems/fundamentals/hello-world/info.json":  `{"id":"hello-world","title":"Hello World","difficulty":"Easy"}`,
		"problems/fundamentals/hello-world/PROBLEM.md": "# Hello World\nReturn a greeting.\n",
		"problems/fundamentals/hello-world/README.md":  "Prefix the name with Hello.\n",
		"problems/fundamentals/hello-world/testcases.json": `[
			{"id":"public","input":{"name":"World"},"expected":"Hello World","hidden":false},
			{"id":"hidden","input":{"name":"Test"},"expected":"Hello Test","hidden":true}
		]`,
		"templates/concept/go/template.go":         "package main\n\ntype Solution struct{}\n",
		"templates/concept/typescript/template.ts": "export class Solution {}\n",
	}
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(root, "problems"), filepath.Join(root, "templates")
}
