package listcmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/cli"
)

type itemRef struct {
	ID    string `json:"id"`
	Order int    `json:"order"`
}

type nodeInfo struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Difficulty string    `json:"difficulty,omitempty"`
	Items      []itemRef `json:"items,omitempty"`
}

func TestListFromLiveTree(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.json"), []byte(`{"concepts_by_id":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	writeNode(t, root, nodeInfo{ID: "root", Title: "Catalog", Items: []itemRef{{ID: "objects", Order: 1}, {ID: "arrays", Order: 2}}})
	for _, category := range []string{"arrays", "objects"} {
		writeNode(t, filepath.Join(root, category), nodeInfo{
			ID: category, Title: category, Items: []itemRef{{ID: "traversal", Order: 1}},
		})
		writeNode(t, filepath.Join(root, category, "traversal"), nodeInfo{
			ID: "traversal", Title: "Traversal", Difficulty: "Easy",
		})
	}
	var output bytes.Buffer
	if err := listFrom(root, "", &output); err != nil {
		t.Fatalf("listFrom: %v", err)
	}
	text := output.String()
	if !reflect.DeepEqual(listedPaths(text), []string{"objects/traversal", "arrays/traversal"}) {
		t.Fatalf("listing order: %s", text)
	}
	for _, path := range []string{"arrays/traversal", "objects/traversal"} {
		id, _ := catalog.ConceptID(path)
		if !strings.Contains(text, path) || !strings.Contains(text, id[:2+catalog.MinConceptIDHexDigits]) || !strings.Contains(text, id) {
			t.Fatalf("missing path or IDs for %s in listing: %s", path, text)
		}
	}
	if !strings.Contains(text, "SHORT ID") || !strings.Contains(text, "FULL ID") {
		t.Fatalf("missing ID columns: %s", text)
	}
	output.Reset()
	writeNode(t, root, nodeInfo{ID: "root", Title: "Catalog", Items: []itemRef{{ID: "objects", Order: 2}, {ID: "arrays", Order: 1}}})
	if err := listFrom(root, "TRAVERSAL", &output); err != nil {
		t.Fatalf("listFrom after reorder: %v", err)
	}
	if !reflect.DeepEqual(listedPaths(output.String()), []string{"arrays/traversal", "objects/traversal"}) {
		t.Fatalf("listing did not follow updated order: %s", output.String())
	}
	output.Reset()
	if err := listFrom(root, "missing", &output); err != nil {
		t.Fatalf("listFrom no matches: %v", err)
	}
	if output.String() != "\nConcepts:\n  No concepts match \"missing\".\n" {
		t.Fatalf("no-match output = %q", output.String())
	}
}

func TestListInterleavesCategoriesAndConceptsByOrder(t *testing.T) {
	root := t.TempDir()
	writeNode(t, root, nodeInfo{ID: "root", Title: "Catalog", Items: []itemRef{{ID: "zeta", Order: 1}, {ID: "alpha", Order: 2}}})
	writeNode(t, filepath.Join(root, "zeta"), nodeInfo{
		ID: "zeta", Title: "Zeta", Items: []itemRef{{ID: "late", Order: 3}, {ID: "nested", Order: 2}, {ID: "early", Order: 1}},
	})
	writeNode(t, filepath.Join(root, "zeta", "early"), nodeInfo{ID: "early", Title: "Early", Difficulty: "Easy"})
	writeNode(t, filepath.Join(root, "zeta", "late"), nodeInfo{ID: "late", Title: "Late", Difficulty: "Easy"})
	writeNode(t, filepath.Join(root, "zeta", "nested"), nodeInfo{
		ID: "nested", Title: "Nested", Items: []itemRef{{ID: "inside", Order: 1}},
	})
	writeNode(t, filepath.Join(root, "zeta", "nested", "inside"), nodeInfo{ID: "inside", Title: "Inside", Difficulty: "Easy"})
	writeNode(t, filepath.Join(root, "alpha"), nodeInfo{
		ID: "alpha", Title: "Alpha", Items: []itemRef{{ID: "last", Order: 1}},
	})
	writeNode(t, filepath.Join(root, "alpha", "last"), nodeInfo{ID: "last", Title: "Last", Difficulty: "Easy"})

	var output bytes.Buffer
	if err := listFrom(root, "", &output); err != nil {
		t.Fatalf("listFrom: %v", err)
	}
	want := []string{"zeta/early", "zeta/nested/inside", "zeta/late", "alpha/last"}
	if got := listedPaths(output.String()); !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %v, want %v\n%s", got, want, output.String())
	}
}

func TestListTextReportsCatalogError(t *testing.T) {
	var output, stderr bytes.Buffer
	err := listFrom(filepath.Join(t.TempDir(), "missing"), "", &output)
	if code := cli.ExitCode(err, &stderr); code != 1 || output.Len() != 0 ||
		!strings.Contains(stderr.String(), "Concepts not listed") || !strings.Contains(stderr.String(), "Reason:") {
		t.Fatalf("exit code = %d, output = %q, stderr = %q", code, output.String(), stderr.String())
	}
}

func listedPaths(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	paths := make([]string, 0, len(lines)-2)
	for _, line := range lines[2:] {
		paths = append(paths, strings.Fields(line)[0])
	}
	return paths
}

func writeNode(t *testing.T, dir string, node nodeInfo) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create node directory: %v", err)
	}
	data, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("marshal node: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "info.json"), data, 0o600); err != nil {
		t.Fatalf("write node metadata: %v", err)
	}
}
