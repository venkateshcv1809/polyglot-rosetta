package catalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type rawItemRef = ItemRef

type rawNodeInfo struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Difficulty  string       `json:"difficulty,omitempty"`
	Description string       `json:"description,omitempty"`
	Items       []rawItemRef `json:"items,omitempty"`
}

func TestBuildCategoryRecursesIntoNestedCategories(t *testing.T) {
	root := t.TempDir()
	writeNode(t, filepath.Join(root, "algorithms"), rawNodeInfo{
		ID:    "algorithms",
		Title: "Algorithms",
		Items: []rawItemRef{{ID: "sorting", Order: 2}, {ID: "search", Order: 1}},
	})
	writeNode(t, filepath.Join(root, "algorithms", "sorting"), rawNodeInfo{
		ID:    "sorting",
		Title: "Sorting",
		Items: []rawItemRef{{ID: "quick-sort", Order: 1}},
	})
	writeNode(t, filepath.Join(root, "algorithms", "sorting", "quick-sort"), rawNodeInfo{
		ID:         "quick-sort",
		Title:      "Quick Sort",
		Difficulty: "Medium",
	})
	writeNode(t, filepath.Join(root, "algorithms", "search"), rawNodeInfo{
		ID:         "search",
		Title:      "Search",
		Difficulty: "Easy",
	})

	category, err := scanCategory(os.DirFS(root), ".", rawItemRef{ID: "algorithms", Order: 1}, make(map[string]Concept))
	if err != nil {
		t.Fatalf("buildCategory: %v", err)
	}
	searchID, _ := ConceptID("algorithms/search")
	if len(category.Concepts) != 1 || category.Concepts[0].ID != searchID {
		t.Fatalf("direct concepts = %+v, want search", category.Concepts)
	}
	if len(category.Categories) != 1 || category.Categories[0].ID != "sorting" {
		t.Fatalf("nested categories = %+v, want sorting", category.Categories)
	}
	quickSortID, _ := ConceptID("algorithms/sorting/quick-sort")
	if got := category.Categories[0].Concepts; len(got) != 1 || got[0].ID != quickSortID {
		t.Fatalf("nested concepts = %+v, want quick-sort", got)
	}
}

func TestBuildCategoryRejectsTraversalID(t *testing.T) {
	root := t.TempDir()
	_, err := scanCategory(os.DirFS(root), ".", rawItemRef{ID: "../outside"}, make(map[string]Concept))
	if err == nil || !strings.Contains(err.Error(), "lowercase kebab-case") {
		t.Fatalf("error = %v, want invalid ID error", err)
	}
}

func TestBuildCategoryRejectsSymlinkedDirectory(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Skipf("create symlink: %v", err)
	}

	_, err := scanCategory(os.DirFS(root), ".", rawItemRef{ID: "escape"}, make(map[string]Concept))
	if err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("error = %v, want symbolic link error", err)
	}
}

func TestBuildCategoryRejectsMismatchedMetadataID(t *testing.T) {
	root := t.TempDir()
	writeNode(t, filepath.Join(root, "algorithms"), rawNodeInfo{ID: "different"})

	_, err := scanCategory(os.DirFS(root), ".", rawItemRef{ID: "algorithms"}, make(map[string]Concept))
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("error = %v, want metadata ID mismatch", err)
	}
}

func TestBuildCategoryAllowsRepeatedLeafIDsAcrossCategories(t *testing.T) {
	root := t.TempDir()
	for _, categoryID := range []string{"one", "two"} {
		writeNode(t, filepath.Join(root, categoryID), rawNodeInfo{
			ID: categoryID, Title: categoryID, Items: []rawItemRef{{ID: "same", Order: 1}},
		})
		writeNode(t, filepath.Join(root, categoryID, "same"), rawNodeInfo{
			ID: "same", Title: "Same", Difficulty: "Easy",
		})
	}

	seen := make(map[string]Concept)
	first, err := scanCategory(os.DirFS(root), ".", rawItemRef{ID: "one"}, seen)
	if err != nil {
		t.Fatalf("first category: %v", err)
	}
	second, err := scanCategory(os.DirFS(root), ".", rawItemRef{ID: "two"}, seen)
	if err != nil {
		t.Fatalf("second category: %v", err)
	}
	firstID, _ := ConceptID("one/same")
	secondID, _ := ConceptID("two/same")
	if first.Concepts[0].ID != firstID || second.Concepts[0].ID != secondID || seen[firstID].Path != "one/same" || seen[secondID].Path != "two/same" {
		t.Fatalf("concept references = %+v, %+v; index = %+v", first.Concepts, second.Concepts, seen)
	}
	if _, err := scanCategory(os.DirFS(root), ".", rawItemRef{ID: "one"}, seen); err == nil || !strings.Contains(err.Error(), "concept ID collision") {
		t.Fatalf("repeated path error = %v", err)
	}
}

func TestCatalogSortingUsesOrderThenID(t *testing.T) {
	refs := sortedRefs([]ItemRef{
		{ID: "later", Order: 2},
		{ID: "same-b", Order: 1},
		{ID: "same-a", Order: 1},
	})

	for index, want := range []string{"same-a", "same-b", "later"} {
		if refs[index].ID != want {
			t.Fatalf("refs[%d].ID = %q, want %q", index, refs[index].ID, want)
		}
	}
}

func writeNode(t *testing.T, dir string, node rawNodeInfo) {
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
