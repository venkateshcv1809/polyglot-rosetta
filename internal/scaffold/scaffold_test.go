package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"polyglot-rosetta/internal/catalog"
)

func TestRegisterInParentUsesMaximumOrderAndSortsItems(t *testing.T) {
	root := t.TempDir()
	infoPath := filepath.Join(root, "info.json")
	writeFixtureFile(t, infoPath, `{
  "id": "root",
  "title": "Root",
  "items": [
    {"id": "later", "order": 10},
    {"id": "first-b", "order": 1},
    {"id": "first-a", "order": 1}
  ]
}`)

	scaffolder := New(root, t.TempDir())
	if err := scaffolder.registerInParent("", "new-item"); err != nil {
		t.Fatalf("registerInParent: %v", err)
	}

	got := readParentInfo(t, infoPath).Items
	want := []catalog.ItemRef{
		{ID: "first-a", Order: 1},
		{ID: "first-b", Order: 1},
		{ID: "later", Order: 10},
		{ID: "new-item", Order: 11},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("items = %+v, want %+v", got, want)
	}

	data, err := os.ReadFile(infoPath)
	if err != nil {
		t.Fatalf("read parent metadata: %v", err)
	}
	if data[len(data)-1] != '\n' {
		t.Fatal("parent metadata does not end with a newline")
	}
}

func TestRenderTemplateEscapesJSONValues(t *testing.T) {
	template := []byte(`{"title":"{{TITLE}}","description":"{{DESCRIPTION}}"}`)
	wantTitle := `Quoted "title"`
	wantDescription := "first line\nsecond line\\path"

	result, err := renderTemplate(template, map[string]string{
		"{{TITLE}}":       wantTitle,
		"{{DESCRIPTION}}": wantDescription,
	}, true)
	if err != nil {
		t.Fatalf("renderTemplate: %v", err)
	}

	var got struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("unmarshal rendered template: %v", err)
	}
	if got.Title != wantTitle || got.Description != wantDescription {
		t.Fatalf("rendered values = %+v, want title %q and description %q", got, wantTitle, wantDescription)
	}
}

func TestRenderTemplatePreservesPlainText(t *testing.T) {
	result, err := renderTemplate([]byte("# {{TITLE}}\n"), map[string]string{"{{TITLE}}": `A "title"`}, false)
	if err != nil {
		t.Fatalf("renderTemplate: %v", err)
	}
	if got, want := string(result), "# A \"title\"\n"; got != want {
		t.Fatalf("result = %q, want %q", got, want)
	}
}
