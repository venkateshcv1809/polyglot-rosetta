package catalogcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	catalogapi "polyglot-rosetta/internal/api/catalog"
	"polyglot-rosetta/internal/cli"
)

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
		{name: "missing URL", args: []string{"--format=json"}, want: "missing required option"},
		{name: "sync error", args: []string{"--url=https://static.example/problems/", "--format=json"}, want: "network unavailable"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output, stderr bytes.Buffer
			err := catalogCommandWithIO(context.Background(), "prosetta catalog", test.args, &output,
				func(context.Context, string, string) (string, error) {
					if test.name == "missing URL" {
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
