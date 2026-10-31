package run

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/domain/evaluation"
	"polyglot-rosetta/internal/toolchain"
)

func TestEvaluateResolvedConcept(t *testing.T) {
	useTestNativeCache(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain is unavailable")
	}
	t.Setenv(catalogEnvironment, writePracticeCatalog(t))
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	concept, err := ResolveConcept(id)
	if err != nil {
		t.Fatal(err)
	}
	source, err := LoadSolution(concept, "go")
	if err != nil {
		t.Fatal(err)
	}
	evaluated, err := Evaluate(context.Background(), runapi.Request{
		ID: id, Language: "go", Source: source, Mode: evaluation.ModePublic,
	}, concept)
	if err != nil || !evaluated.Report.PassedAll || evaluated.Report.TotalTests != 1 {
		t.Fatalf("evaluation = %+v, error = %v", evaluated, err)
	}
	otherID, _ := catalog.ConceptID("objects/traversal")
	if _, err := Evaluate(context.Background(), runapi.Request{
		ID: otherID, Language: "go", Source: source, Mode: evaluation.ModePublic,
	}, concept); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("mismatched concept error = %v", err)
	}
}

func TestEvaluateStopsBeforeLoadingCasesWhenToolchainIsUnavailable(t *testing.T) {
	cachePath := useTestNativeCache(t).Toolchain
	if _, err := toolchain.RefreshCache(cachePath,
		func(string) (string, error) { return "", errors.New("not found") },
		func(string, string) (string, error) { t.Fatal("unexpected inspect"); return "", nil },
	); err != nil {
		t.Fatal(err)
	}
	t.Setenv(catalogEnvironment, writePracticeCatalog(t))
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	concept, err := ResolveConcept(id)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Evaluate(context.Background(), runapi.Request{
		ID: id, Language: "go", Source: "invalid source", Mode: evaluation.ModePublic,
	}, concept)
	if !errors.Is(err, toolchain.ErrUnavailable) {
		t.Fatalf("error = %v, want cached unavailability", err)
	}
}
