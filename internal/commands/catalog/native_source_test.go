package catalogcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	catalogapi "polyglot-rosetta/internal/api/catalog"
	"polyglot-rosetta/internal/version"
)

func TestCatalogExplicitLocalSourceWins(t *testing.T) {
	previous := version.Value
	t.Cleanup(func() { version.Value = previous })
	for _, build := range []string{"dev", "v0.1.0"} {
		t.Run(build, func(t *testing.T) {
			version.Value = build
			useTestCatalogCache(t)
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "info.json"), []byte(`{"id":"root","title":"Local catalog","items":[]}`), 0o600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PROSETTA_CATALOG", root)
			t.Setenv("PROSETTA_CATALOG_URL", "https://environment.example/catalog/")
			for _, args := range [][]string{
				{"--format=json"},
				{"--format=json", "--url=https://flag.example/catalog/"},
			} {
				var output bytes.Buffer
				err := catalogCommandWithIO(context.Background(), "prosetta catalog", args, &output,
					func(context.Context, string, string) (string, error) {
						t.Fatal("explicit local source attempted a download")
						return "", nil
					})
				if err != nil {
					t.Fatal(err)
				}
				var result catalogapi.Result
				if err := json.Unmarshal(output.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				if result.Status != catalogapi.StatusOK || result.Root != root {
					t.Fatalf("result = %+v", result)
				}
			}
		})
	}
}

func TestCatalogInvalidExplicitLocalSourceDoesNotDownload(t *testing.T) {
	previous := version.Value
	version.Value = "v0.1.0"
	t.Cleanup(func() { version.Value = previous })
	useTestCatalogCache(t)
	t.Setenv("PROSETTA_CATALOG", filepath.Join(t.TempDir(), "missing"))
	var output bytes.Buffer
	err := catalogCommandWithIO(context.Background(), "prosetta catalog", []string{"--format=json"}, &output,
		func(context.Context, string, string) (string, error) {
			t.Fatal("invalid explicit local source fell back to downloading")
			return "", nil
		})
	if err == nil {
		t.Fatal("invalid local source accepted")
	}
	var result catalogapi.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != catalogapi.StatusError || result.Error == "" {
		t.Fatalf("result = %+v", result)
	}
}
