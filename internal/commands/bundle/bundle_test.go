package bundlecmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polyglot-rosetta/internal/cli"
)

func TestBundleFromCreatesArtifactsAndReportsPath(t *testing.T) {
	outputDir := t.TempDir()
	var output bytes.Buffer
	if err := bundleFrom("../../../problems", "../../../templates", outputDir, &output); err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "\nOK:\n  Bundle created\n  Path:      "+outputDir+"\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
	for _, name := range []string{"index.json", "manifest.json", "templates.tar.gz", "fundamentals.tar.gz"} {
		if _, err := os.Stat(filepath.Join(outputDir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}

func TestBundleFromFormatsBuildFailure(t *testing.T) {
	var output, stderr bytes.Buffer
	err := bundleFrom(filepath.Join(t.TempDir(), "missing"), "../../../templates", t.TempDir(), &output)
	if code := cli.ExitCode(err, &stderr); code != 1 || output.Len() != 0 ||
		!strings.Contains(stderr.String(), "Bundle not created") || !strings.Contains(stderr.String(), "Reason:") {
		t.Fatalf("exit code = %d, stdout = %q, stderr = %q", code, output.String(), stderr.String())
	}
}
