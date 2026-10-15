package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polyglot-rosetta/internal/domain/catalog"
	"polyglot-rosetta/internal/domain/language"
)

func TestCreateCategoryRegistersRenderedCategory(t *testing.T) {
	root, templates := scaffoldFixture(t)
	scaffolder := New(root, templates)
	options := CreateCategoryOptions{
		ID: "algorithms", Title: "Algorithms", Description: "Algorithm practice",
	}
	if err := scaffolder.CreateCategory(options); err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	category := readParentInfo(t, filepath.Join(root, "algorithms", "info.json"))
	if category.ID != options.ID || category.Title != options.Title || category.Description != options.Description {
		t.Fatalf("category = %+v", category)
	}
	rootInfo := readParentInfo(t, filepath.Join(root, "info.json"))
	if len(rootInfo.Items) != 1 || rootInfo.Items[0] != (catalog.ItemRef{ID: options.ID, Order: 1}) {
		t.Fatalf("root items = %+v", rootInfo.Items)
	}
}

func TestCreateConceptCopiesReferenceSourceWithoutPracticeStarter(t *testing.T) {
	root, templates := scaffoldFixture(t)
	scaffolder := New(root, templates)
	options := CreateConceptOptions{
		ID: "hello-world", Title: "Hello World", Difficulty: catalog.DifficultyEasy,
	}
	if err := scaffolder.CreateConcept(options); err != nil {
		t.Fatalf("CreateConcept: %v", err)
	}

	conceptDir := filepath.Join(root, options.ID)
	for _, lang := range language.AllSupported() {
		mainSource, err := os.ReadFile(filepath.Join(conceptDir, targetRelPath(lang)))
		if err != nil {
			t.Errorf("main source for %s: %v", lang, err)
			continue
		}
		if _, err := os.Stat(filepath.Join(conceptDir, templateRelPath(lang)+".txt")); !os.IsNotExist(err) {
			t.Errorf("practice starter for %s should not be generated: %v", lang, err)
		}
		if string(mainSource) == "" {
			t.Errorf("reference source for %s is empty", lang)
		}
	}
	var concept catalog.ConceptInfo
	readJSON(t, filepath.Join(conceptDir, "info.json"), &concept)
	if concept.ID != options.ID || concept.Difficulty != options.Difficulty {
		t.Fatalf("concept = %+v", concept)
	}
}

func TestCreateConceptAllowsSameLeafIDUnderDifferentParents(t *testing.T) {
	root, templates := scaffoldFixture(t)
	scaffolder := New(root, templates)
	for _, parent := range []string{"arrays", "objects"} {
		if err := scaffolder.CreateCategory(CreateCategoryOptions{ID: parent, Title: parent}); err != nil {
			t.Fatalf("CreateCategory(%q): %v", parent, err)
		}
		if err := scaffolder.CreateConcept(CreateConceptOptions{
			ParentPath: parent, ID: "traversal", Title: "Traversal", Difficulty: catalog.DifficultyEasy,
		}); err != nil {
			t.Fatalf("CreateConcept(%q): %v", parent, err)
		}
		if _, err := os.Stat(filepath.Join(root, parent, "traversal", "info.json")); err != nil {
			t.Fatalf("concept under %q: %v", parent, err)
		}
	}
	if err := scaffolder.CreateConcept(CreateConceptOptions{
		ParentPath: "arrays", ID: "traversal", Title: "Traversal", Difficulty: catalog.DifficultyEasy,
	}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("duplicate path error = %v", err)
	}
}

func TestCreateCategoryRejectsParentOutsideCatalog(t *testing.T) {
	root, templates := scaffoldFixture(t)
	scaffolder := New(root, templates)

	for _, parent := range []string{"../outside", filepath.Join(root, "absolute")} {
		err := scaffolder.CreateCategory(CreateCategoryOptions{ParentPath: parent, ID: "escape", Title: "Escape"})
		if err == nil || !strings.Contains(err.Error(), "catalog") {
			t.Errorf("CreateCategory parent %q error = %v, want containment error", parent, err)
		}
	}
}

func TestCreateCategoryRejectsSymlinkOutsideCatalog(t *testing.T) {
	root, templates := scaffoldFixture(t)
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "outside")); err != nil {
		t.Skipf("create symlink: %v", err)
	}

	err := New(root, templates).CreateCategory(CreateCategoryOptions{
		ParentPath: "outside", ID: "escape", Title: "Escape",
	})
	if err == nil || !strings.Contains(err.Error(), "outside the catalog") {
		t.Fatalf("CreateCategory error = %v, want symlink containment error", err)
	}
}

func TestCreateCategoryRemovesStagingDirectoryAfterTemplateFailure(t *testing.T) {
	root, templates := scaffoldFixture(t)
	templatePath := filepath.Join(templates, "category", "info.json")
	if err := os.Remove(templatePath); err != nil {
		t.Fatalf("remove category template: %v", err)
	}
	if err := os.Mkdir(templatePath, 0o755); err != nil {
		t.Fatalf("replace category template with directory: %v", err)
	}

	err := New(root, templates).CreateCategory(CreateCategoryOptions{ID: "broken", Title: "Broken"})
	if err == nil {
		t.Fatal("CreateCategory returned nil error")
	}
	if _, statErr := os.Stat(filepath.Join(root, "broken")); !os.IsNotExist(statErr) {
		t.Fatalf("target exists after failure: %v", statErr)
	}
	entries, readErr := os.ReadDir(root)
	if readErr != nil {
		t.Fatalf("read catalog root: %v", readErr)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".broken-") {
			t.Fatalf("staging directory %q remains after failure", entry.Name())
		}
	}
}

func TestCreateCategoryRejectsMalformedParentMetadataWithoutCreatingTarget(t *testing.T) {
	root, templates := scaffoldFixture(t)
	writeFixtureFile(t, filepath.Join(root, "info.json"), `{not-json`)

	err := New(root, templates).CreateCategory(CreateCategoryOptions{ID: "orphan", Title: "Orphan"})
	if err == nil || !strings.Contains(err.Error(), "failed to parse parent") {
		t.Fatalf("CreateCategory error = %v, want parent parsing error", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "orphan")); !os.IsNotExist(statErr) {
		t.Fatalf("target exists after failure: %v", statErr)
	}
}

func scaffoldFixture(t *testing.T) (string, string) {
	t.Helper()
	base := t.TempDir()
	root := filepath.Join(base, "problems")
	templates := filepath.Join(base, "templates")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("create root: %v", err)
	}
	writeFixtureFile(t, filepath.Join(root, "info.json"), `{"id":"root","title":"Root","items":[]}`)
	writeFixtureFile(t, filepath.Join(templates, "category", "info.json"), `{"id":"{{ID}}","title":"{{TITLE}}","description":"{{DESCRIPTION}}","items":[]}`)
	writeFixtureFile(t, filepath.Join(templates, "concept", "info.json"), `{"id":"{{ID}}","title":"{{TITLE}}","difficulty":"{{DIFFICULTY}}","description":"{{DESCRIPTION}}"}`)
	writeFixtureFile(t, filepath.Join(templates, "concept", "testcases.json"), `[]`)
	writeFixtureFile(t, filepath.Join(templates, "concept", "README.md"), "# {{TITLE}}\n")
	writeFixtureFile(t, filepath.Join(templates, "concept", "PROBLEM.md"), "# {{TITLE}}\n")
	for _, lang := range language.AllSupported() {
		writeFixtureFile(t, filepath.Join(templates, "concept", templateRelPath(lang)), "{{TITLE}}\n")
	}
	return root, templates
}

func writeFixtureFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}

func readParentInfo(t *testing.T, path string) catalog.ParentInfo {
	t.Helper()
	var info catalog.ParentInfo
	readJSON(t, path, &info)
	return info
}

func readJSON(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}
