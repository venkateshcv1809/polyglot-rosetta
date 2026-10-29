package bundle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"polyglot-rosetta/internal/version"
)

func TestBuildMinimumEngine(t *testing.T) {
	repo, templates, output := t.TempDir(), t.TempDir(), t.TempDir()
	root := filepath.Join(repo, "problems")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	info := []byte(`{"id":"root","title":"Test catalog","items":[]}`)
	infoPath := filepath.Join(root, "info.json")
	if err := os.WriteFile(infoPath, info, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Build(root, templates, output); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(output, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.MinEngine != version.MinimumEngine() {
		t.Fatalf("minimum engine = %q, want %q", manifest.MinEngine, version.MinimumEngine())
	}
	unchanged, err := os.ReadFile(infoPath)
	if err != nil || string(unchanged) != string(info) {
		t.Fatalf("root info.json changed: %s, error = %v", unchanged, err)
	}
}
