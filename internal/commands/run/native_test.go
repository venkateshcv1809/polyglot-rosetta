package runcmd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/domain/evaluation"
	runservice "polyglot-rosetta/internal/run"
	"polyglot-rosetta/internal/toolchain"
)

func TestNativeHostChecksResolvedConceptIdentity(t *testing.T) {
	t.Setenv("PROSETTA_CATALOG", writeCatalog(t))
	id := mustConceptID(t, "fundamentals/hello-world")
	concept, err := runservice.ResolveConcept(id)
	if err != nil {
		t.Fatal(err)
	}
	host := nativeHost{concept: concept}
	loaded, err := host.Resolve(context.Background(), strings.ToUpper(id[:10]))
	if err != nil || loaded.Path != concept.Reference || len(loaded.TestCases) != 1 {
		t.Fatalf("Resolve = %+v, %v", loaded, err)
	}
	wrong := mustConceptID(t, "fundamentals/another-concept")
	if _, err := host.Resolve(context.Background(), wrong); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("mismatched identity error = %v", err)
	}
}

func TestNativeHostRunsSelectedCasesThroughSharedCore(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain is unavailable")
	}
	cache := useRunTestCache(t).Toolchain
	if _, err := toolchain.RefreshCache(cache,
		func(name string) (string, error) { return "/tools/" + name, nil },
		func(string, string) (string, error) { return "fixture", nil },
	); err != nil {
		t.Fatal(err)
	}
	root := writeCatalog(t)
	t.Setenv("PROSETTA_CATALOG", root)
	cases := `[{"id":"public","input":{"name":"World"},"expected":"Hello World"},{"id":"hidden","input":{"name":"World"},"expected":"different","hidden":true}]`
	if err := os.WriteFile(filepath.Join(root, "fundamentals", "hello-world", "testcases.json"), []byte(cases), 0o600); err != nil {
		t.Fatal(err)
	}
	id := mustConceptID(t, "fundamentals/hello-world")
	concept, err := runservice.ResolveConcept(id)
	if err != nil {
		t.Fatal(err)
	}
	source, err := runservice.LoadSolution(concept, "go")
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []evaluation.Mode{evaluation.ModePublic, evaluation.ModeAll} {
		result, err := Run(context.Background(), runapi.Request{ID: id[:10], Language: "go", Source: source, Mode: mode}, nativeHost{concept: concept})
		if err != nil || result.ID != id || result.TotalPassed != 1 {
			t.Fatalf("Run(%q) = %+v, %v", mode, result, err)
		}
		if mode == evaluation.ModePublic && (result.Status != evaluation.StatusPassed || len(result.TestResults) != 1) {
			t.Fatalf("public result = %+v", result)
		}
		if mode == evaluation.ModeAll && (result.Status != evaluation.StatusFailed || result.TotalFailed != 1 || !result.TestResults[1].Hidden) {
			t.Fatalf("all result = %+v", result)
		}
	}
}
