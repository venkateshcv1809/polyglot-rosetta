package runcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/domain/evaluation"
	"polyglot-rosetta/internal/toolchain"
)

func TestSourceCodeReadsEachInputMode(t *testing.T) {
	tests := []struct {
		name  string
		opts  options
		stdin string
		want  string
	}{
		{name: "stdin", opts: options{stdin: true}, stdin: "stream source", want: "stream source"},
		{name: "inline", opts: options{code: "inline source"}, want: "inline source"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := sourceCode(test.opts, strings.NewReader(test.stdin))
			if err != nil {
				t.Fatalf("sourceCode: %v", err)
			}
			if got == nil || *got != test.want {
				t.Fatalf("source = %v, want %q", got, test.want)
			}
		})
	}
}

func TestSourceCodeUsesCatalogByDefault(t *testing.T) {
	source, err := sourceCode(options{}, strings.NewReader(""))
	if err != nil || source != nil {
		t.Fatalf("source = %v, error = %v", source, err)
	}
}

func TestRunUsesActivatedCatalog(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG", "")
	t.Chdir(t.TempDir())
	cache := useRunTestCache(t).Catalog
	version := strings.Repeat("a", 64)
	root := filepath.Join(cache, "versions", version, "problems")
	conceptDir := filepath.Join(root, "fundamentals", "hello-world")
	if err := os.MkdirAll(conceptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCatalogInfo(t, root, "fundamentals", "hello-world")
	if err := os.WriteFile(filepath.Join(conceptDir, "testcases.json"), []byte(`[{"id":1,"input":{},"expected":true}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cache, "current"), []byte(version+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	if _, err := exec.LookPath("go"); err == nil {
		var output bytes.Buffer
		err := runWithIO(context.Background(), options{
			id: id, language: "go",
			code: "package main\ntype Solution struct{}\nfunc (Solution) Solve(any) any { return true }\n",
		}, strings.NewReader(""), &output)
		if err != nil || !strings.Contains(output.String(), "1/1 tests passed") {
			t.Fatalf("run cached concept: %v\n%s", err, output.String())
		}
	}
}

func TestRunStopsBeforeExecutionWhenToolchainIsCachedUnavailable(t *testing.T) {
	cachePath := useRunTestCache(t).Toolchain
	if _, err := toolchain.RefreshCache(cachePath,
		func(string) (string, error) { return "", errors.New("not found") },
		func(string, string) (string, error) { t.Fatal("unexpected inspect"); return "", nil },
	); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PROSETTA_CATALOG", writeCatalog(t))
	id := mustConceptID(t, "fundamentals/hello-world")
	var output bytes.Buffer
	err := runWithIO(context.Background(), options{
		id: id, language: "go", code: "invalid source",
	}, strings.NewReader(""), &output)
	if !errors.Is(err, toolchain.ErrUnavailable) || strings.Contains(output.String(), "Executing") {
		t.Fatalf("error = %v, output = %q", err, output.String())
	}
	output.Reset()
	err = runWithIO(context.Background(), options{
		id: id, language: "go", code: "invalid source", format: "json",
	}, strings.NewReader(""), &output)
	if !errors.Is(err, toolchain.ErrUnavailable) {
		t.Fatalf("JSON run error = %v", err)
	}
	var result runapi.Result
	if decodeErr := json.Unmarshal(output.Bytes(), &result); decodeErr != nil ||
		result.Status != evaluation.StatusError || !strings.Contains(result.Error, "language toolchain unavailable") {
		t.Fatalf("JSON result = %+v, decode error = %v", result, decodeErr)
	}
}

func TestRunResolvesRepeatedLeafIDsByHashedPath(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain is unavailable")
	}
	root := t.TempDir()
	for _, category := range []string{"arrays", "objects"} {
		dir := filepath.Join(root, category, "traversal")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create concept: %v", err)
		}
		data := `[{"id":1,"input":{},"expected":"` + category + `","hidden":false}]`
		if err := os.WriteFile(filepath.Join(dir, "testcases.json"), []byte(data), 0o600); err != nil {
			t.Fatalf("write test cases: %v", err)
		}
		writeCatalogInfo(t, root, category, "traversal")
	}
	if err := os.WriteFile(filepath.Join(root, "info.json"), []byte(`{"id":"root","title":"Problems","items":[{"id":"arrays","order":1},{"id":"objects","order":2}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PROSETTA_CATALOG", root)
	source := "package main\ntype Solution struct{}\nfunc (Solution) Solve(any) any { return \"arrays\" }\n"
	for _, test := range []struct {
		ref      string
		wantFail bool
	}{
		{ref: "arrays/traversal"},
		{ref: "objects/traversal", wantFail: true},
	} {
		id, _ := catalog.ConceptID(test.ref)
		var output bytes.Buffer
		err := runWithIO(context.Background(), options{
			id: id, language: "go", code: source,
		}, strings.NewReader(""), &output)
		if (err != nil) != test.wantFail || (test.wantFail && !errors.Is(err, errTestsFailed)) {
			t.Fatalf("runWithIO(%q) error = %v, wantFail %t", test.ref, err, test.wantFail)
		}
		if !strings.Contains(output.String(), "Summary:\n  Concept:      "+test.ref) {
			t.Fatalf("summary for %q: %s", test.ref, output.String())
		}
		if strings.Contains(output.String(), "Run:") || !strings.Contains(output.String(), "Cases:        public") {
			t.Fatalf("run layout for %q: %s", test.ref, output.String())
		}
	}
}

func TestRunEvaluatesAllSourceModes(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain is unavailable")
	}
	t.Setenv("PROSETTA_CATALOG", writeCatalog(t))
	source := `package main
type Solution struct{}
func (Solution) Solve(input any) any { return "Hello " + input.(map[string]any)["name"].(string) }
`
	tests := []struct {
		name  string
		opts  options
		stdin string
	}{
		{name: "catalog", opts: options{}},
		{name: "stdin", opts: options{stdin: true}, stdin: source},
		{name: "code", opts: options{code: source}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			test.opts.id, _ = catalog.ConceptID("fundamentals/hello-world")
			test.opts.language = "go"
			err := runWithIO(context.Background(), test.opts, strings.NewReader(test.stdin), &output)
			if err != nil {
				t.Fatalf("runWithIO: %v\n%s", err, output.String())
			}
			if !strings.Contains(output.String(), "[ PASS ]") || !strings.Contains(output.String(), "1/1 tests passed") {
				t.Fatalf("output:\n%s", output.String())
			}
		})
	}
}

func TestRunReturnsFailureForWrongAnswer(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain is unavailable")
	}
	t.Setenv("PROSETTA_CATALOG", writeCatalog(t))
	err := runWithIO(context.Background(), options{
		id: mustConceptID(t, "fundamentals/hello-world"), language: "go",
		code: "package main\ntype Solution struct{}\nfunc (Solution) Solve(any) any { return \"wrong\" }\n",
	}, strings.NewReader(""), &bytes.Buffer{})
	if !errors.Is(err, errTestsFailed) {
		t.Fatalf("error = %v, want errTestsFailed", err)
	}
}

func TestRunJSONOutput(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain is unavailable")
	}
	t.Setenv("PROSETTA_CATALOG", writeCatalog(t))
	id := mustConceptID(t, "fundamentals/hello-world")
	for _, test := range []struct {
		name       string
		code       string
		wantStatus evaluation.Status
		wantCase   evaluation.TestStatus
		wantError  bool
	}{
		{name: "passed", code: `package main
type Solution struct{}
func (Solution) Solve(input any) any { return "Hello " + input.(map[string]any)["name"].(string) }
`, wantStatus: evaluation.StatusPassed, wantCase: evaluation.TestPassed},
		{name: "failed", code: `package main
type Solution struct{}
func (Solution) Solve(any) any { return "wrong" }
`, wantStatus: evaluation.StatusFailed, wantCase: evaluation.TestWrongAnswer, wantError: true},
		{name: "crashed", code: `package main
type Solution struct{}
func (Solution) Solve(any) any { panic("boom") }
`, wantStatus: evaluation.StatusFailed, wantCase: evaluation.TestCrashed, wantError: true},
		{name: "execution error", code: "not valid Go", wantStatus: evaluation.StatusError, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			err := runWithIO(context.Background(), options{
				id: id, language: "go", code: test.code, format: "json",
			}, strings.NewReader(""), &output)
			if (err != nil) != test.wantError {
				t.Fatalf("run error = %v, want error %t", err, test.wantError)
			}
			var result runapi.Result
			decoder := json.NewDecoder(&output)
			if err := decoder.Decode(&result); err != nil {
				t.Fatalf("decode JSON: %v; output = %q", err, output.String())
			}
			if decoder.More() || output.Len() != 0 {
				t.Fatalf("unexpected trailing output: %q", output.String())
			}
			if result.ID != id || result.Concept != "fundamentals/hello-world" || result.Language != "go" || result.Mode != evaluation.ModePublic || result.InputSource != "inline" || result.Status != test.wantStatus {
				t.Fatalf("result = %+v", result)
			}
			if test.wantStatus == evaluation.StatusError {
				if result.Error == "" || len(result.TestResults) != 0 {
					t.Fatalf("error result = %+v", result)
				}
				return
			}
			if result.TotalPassed+result.TotalFailed != 1 || len(result.TestResults) != 1 ||
				result.TestResults[0].Status != test.wantCase ||
				result.TestResults[0].Passed != (test.wantStatus == evaluation.StatusPassed) {
				t.Fatalf("test results = %+v", result)
			}
			if test.wantCase != evaluation.TestCrashed && result.TestResults[0].SolutionTimeNs <= 0 {
				t.Fatalf("missing per-test solution time: %+v", result.TestResults[0])
			}
			if test.wantCase == evaluation.TestCrashed && (result.TestResults[0].SolutionTimeNs != 0 || result.TestResults[0].Message == "") {
				t.Fatalf("crash result = %+v", result.TestResults[0])
			}
			if result.Error != "" {
				t.Fatalf("unexpected error field: %q", result.Error)
			}
		})
	}
}

func TestRunJSONFromStdinIncludesMode(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain is unavailable")
	}
	t.Setenv("PROSETTA_CATALOG", writeCatalog(t))
	source := `package main
type Solution struct{}
func (Solution) Solve(input any) any { return "Hello " + input.(map[string]any)["name"].(string) }
`
	var output bytes.Buffer
	id := mustConceptID(t, "fundamentals/hello-world")
	err := runWithIO(context.Background(), options{
		id: id[:10], language: "go",
		stdin: true, all: true, format: "json",
	}, strings.NewReader(source), &output)
	if err != nil {
		t.Fatal(err)
	}
	var result runapi.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode JSON: %v; output = %q", err, output.String())
	}
	if result.ID != id || result.InputSource != "stdin" || result.Mode != evaluation.ModeAll || result.Status != evaluation.StatusPassed {
		t.Fatalf("result = %+v", result)
	}
}

func TestRunJSONReportsPreparationError(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG", "")
	useRunTestCache(t)
	t.Chdir(t.TempDir())
	var output bytes.Buffer
	err := runWithIO(context.Background(), options{
		id: mustConceptID(t, "fundamentals/hello-world"), language: "go", format: "json",
	}, strings.NewReader(""), &output)
	if err == nil {
		t.Fatal("expected missing catalog error")
	}
	var result runapi.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode JSON: %v; output = %q", err, output.String())
	}
	if result.Status != evaluation.StatusError ||
		!strings.Contains(result.Error, "no catalog is available") ||
		!strings.Contains(result.Error, "prosetta catalog --url") {
		t.Fatalf("result = %+v", result)
	}
}

func TestRunCommandJSONExecutionErrorHasNoTextFailure(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG", "")
	useRunTestCache(t)
	t.Chdir(t.TempDir())
	id := mustConceptID(t, "fundamentals/hello-world")
	var stdout, stderr bytes.Buffer
	err := runCommandWithIO(context.Background(), "prosetta run", []string{
		"--id=" + id, "--lang=go", "--format=json",
	}, strings.NewReader(""), &stdout)
	if code := cli.ExitCode(err, &stderr); code != 1 || stderr.Len() != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	var result runapi.Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode JSON: %v; output = %q", err, stdout.String())
	}
	if result.Status != evaluation.StatusError || !strings.Contains(result.Error, "no catalog is available") {
		t.Fatalf("result = %+v", result)
	}
}

func TestRunSuggestsExplicitSourceWhenCatalogSolutionIsMissing(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG", writeCatalog(t))
	var output bytes.Buffer
	err := runWithIO(context.Background(), options{
		id: mustConceptID(t, "fundamentals/hello-world"), language: "python", format: "json",
	}, strings.NewReader(""), &output)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("error = %v, want missing solution", err)
	}
	var result runapi.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Concept != "fundamentals/hello-world" || !strings.Contains(result.Error, "use --stdin or --code") {
		t.Fatalf("result = %+v", result)
	}
}

func mustConceptID(t *testing.T, path string) string {
	t.Helper()
	id, err := catalog.ConceptID(path)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func writeCatalog(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	conceptDir := filepath.Join(root, "fundamentals", "hello-world")
	if err := os.MkdirAll(conceptDir, 0o755); err != nil {
		t.Fatalf("create concept: %v", err)
	}
	data := `[{"id":1,"input":{"name":"World"},"expected":"Hello World","hidden":false}]`
	if err := os.WriteFile(filepath.Join(conceptDir, "testcases.json"), []byte(data), 0o600); err != nil {
		t.Fatalf("write test cases: %v", err)
	}
	goDir := filepath.Join(conceptDir, "go")
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		t.Fatalf("create go solution directory: %v", err)
	}
	source := "package main\ntype Solution struct{}\nfunc (Solution) Solve(input any) any { return \"Hello \" + input.(map[string]any)[\"name\"].(string) }\n"
	if err := os.WriteFile(filepath.Join(goDir, "main.go"), []byte(source), 0o600); err != nil {
		t.Fatalf("write go solution: %v", err)
	}
	writeCatalogInfo(t, root, "fundamentals", "hello-world")
	return root
}

func writeCatalogInfo(t *testing.T, root, category, concept string) {
	t.Helper()
	for name, data := range map[string]string{
		"info.json":                             `{"id":"root","title":"Problems","items":[{"id":"` + category + `","order":1}]}`,
		category + "/info.json":                 `{"id":"` + category + `","title":"` + category + `","items":[{"id":"` + concept + `","order":1}]}`,
		category + "/" + concept + "/info.json": `{"id":"` + concept + `","title":"` + concept + `","difficulty":"Easy"}`,
	} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(name)), []byte(data), 0o600); err != nil {
			t.Fatalf("write catalog metadata %s: %v", name, err)
		}
	}
}
