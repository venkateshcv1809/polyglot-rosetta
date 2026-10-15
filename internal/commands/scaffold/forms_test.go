package scaffoldcmd

import (
	"testing"

	"polyglot-rosetta/internal/cli"
)

func TestFormSpecsUseStableFieldOrder(t *testing.T) {
	assertFieldOrder(t, "category", CategoryFormSpec, []string{"id", "title", "description", "parent"})
	assertFieldOrder(t, "concept", ConceptFormSpec, []string{"parent", "id", "title", "description", "difficulty"})
}

func assertFieldOrder(t *testing.T, name string, fields []cli.PromptField, want []string) {
	t.Helper()
	if len(fields) != len(want) {
		t.Fatalf("%s fields = %d, want %d", name, len(fields), len(want))
	}
	for index, key := range want {
		if fields[index].Key != key {
			t.Fatalf("%s field %d = %q, want %q", name, index, fields[index].Key, key)
		}
	}
}
