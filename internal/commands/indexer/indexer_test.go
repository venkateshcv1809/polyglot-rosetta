package indexcmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

	category, err := buildCategory(root, root, rawItemRef{ID: "algorithms", Order: 1})
	if err != nil {
		t.Fatalf("buildCategory: %v", err)
	}
	if len(category.Concepts) != 1 || category.Concepts[0].ID != "search" {
		t.Fatalf("direct concepts = %+v, want search", category.Concepts)
	}
	if len(category.Categories) != 1 || category.Categories[0].ID != "sorting" {
		t.Fatalf("nested categories = %+v, want sorting", category.Categories)
	}
	if got := category.Categories[0].Concepts; len(got) != 1 || got[0].ID != "quick-sort" {
		t.Fatalf("nested concepts = %+v, want quick-sort", got)
	}
}

func TestBuildCategoryRejectsTraversalID(t *testing.T) {
	root := t.TempDir()
	_, err := buildCategory(root, root, rawItemRef{ID: "../outside"})
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

	_, err := buildCategory(root, root, rawItemRef{ID: "escape"})
	if err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("error = %v, want symbolic link error", err)
	}
}

func TestBuildCategoryRejectsMismatchedMetadataID(t *testing.T) {
	root := t.TempDir()
	writeNode(t, filepath.Join(root, "algorithms"), rawNodeInfo{ID: "different"})

	_, err := buildCategory(root, root, rawItemRef{ID: "algorithms"})
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("error = %v, want metadata ID mismatch", err)
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
