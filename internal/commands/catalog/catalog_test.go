package catalogcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	catalogapi "polyglot-rosetta/internal/api/catalog"
	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/version"
)

func TestCatalogCommandUsesLocalTreeWhenDev(t *testing.T) {
	previous := version.Value
	version.Value = "dev"
	t.Cleanup(func() { version.Value = previous })
	t.Setenv("PROSETTA_CATALOG_URL", "")
	useTestCatalogCache(t)
	root := filepath.Join(t.TempDir(), "problems")
	if err := os.MkdirAll(filepath.Join(root, "fundamentals", "hello-world"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string]string{
		"info.json":                          `{"id":"root","title":"Problems","items":[{"id":"fundamentals","order":1}]}`,
		"fundamentals/info.json":             `{"id":"fundamentals","title":"Fundamentals","items":[{"id":"hello-world","order":1}]}`,
		"fundamentals/hello-world/info.json": `{"id":"hello-world","title":"Hello World","difficulty":"Easy"}`,
	} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PROSETTA_CATALOG", root)
	var output bytes.Buffer
	err := catalogCommandWithIO(context.Background(), "prosetta catalog", nil, &output,
		func(context.Context, string, string) (string, error) {
			t.Fatal("development catalog fetched a URL")
			return "", nil
		})
	if err != nil || !strings.Contains(output.String(), root) {
		t.Fatalf("output = %q, error = %v", output.String(), err)
	}
}

func TestCatalogCommandTextSuccess(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG_URL", "")
	cacheDir := useTestCatalogCache(t)
	var output bytes.Buffer
	err := catalogCommandWithIO(context.Background(), "prosetta catalog", []string{
		"--url=https://static.example/problems/",
	}, &output, func(_ context.Context, url, cache string) (string, error) {
		if url != "https://static.example/problems/" || cache != cacheDir {
			t.Fatalf("sync args = %q, %q", url, cache)
		}
		return "/tmp/catalog/versions/abc/problems", nil
	})
	if err != nil || output.String() != "\nOK:\n  Catalog ready\n  Path:      /tmp/catalog/versions/abc/problems\n" {
		t.Fatalf("output = %q, error = %v", output.String(), err)
	}
}

func TestCatalogCommandJSONSuccess(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG_URL", "")
	useTestCatalogCache(t)
	var output bytes.Buffer
	err := catalogCommandWithIO(context.Background(), "prosetta catalog", []string{
		"--url=https://static.example/problems/", "--format=json",
	}, &output, func(context.Context, string, string) (string, error) {
		return "/cache/problems", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var result catalogapi.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != catalogapi.StatusOK || result.Root != "/cache/problems" || result.Error != "" {
		t.Fatalf("result = %+v", result)
	}
}

func TestCatalogCommandJSONFailures(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG_URL", "")
	useTestCatalogCache(t)
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{name: "unknown flag", args: []string{"--format=json", "--unknown"}, want: "flag provided but not defined"},
		{name: "sync error", args: []string{"--url=https://static.example/problems/", "--format=json"}, want: "network unavailable"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output, stderr bytes.Buffer
			err := catalogCommandWithIO(context.Background(), "prosetta catalog", test.args, &output,
				func(context.Context, string, string) (string, error) {
					if test.name == "unknown flag" {
						t.Fatal("sync called after parse failure")
					}
					return "", errors.New("network unavailable")
				})
			if code := cli.ExitCode(err, &stderr); code != 1 || stderr.Len() != 0 {
				t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
			}
			var result catalogapi.Result
			if err := json.Unmarshal(output.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Status != catalogapi.StatusError || result.Root != "" || !strings.Contains(result.Error, test.want) {
				t.Fatalf("result = %+v", result)
			}
		})
	}
}

func TestCatalogCommandTextFailureUsesSharedTemplate(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG_URL", "")
	useTestCatalogCache(t)
	var output, stderr bytes.Buffer
	err := catalogCommandWithIO(context.Background(), "prosetta catalog", []string{
		"--url=https://static.example/problems/",
	}, &output, func(context.Context, string, string) (string, error) {
		return "", errors.New("network unavailable")
	})
	if code := cli.ExitCode(err, &stderr); code != 1 || output.Len() != 0 ||
		stderr.String() != "\nERROR:\n  Catalog not ready\n  Reason:  network unavailable\n" {
		t.Fatalf("exit code = %d, stdout = %q, stderr = %q", code, output.String(), stderr.String())
	}
}
