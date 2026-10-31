package runcmd

import (
	"context"
	"fmt"
	"strings"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/domain/evaluation"
	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/domain/testcase"
	engine "polyglot-rosetta/internal/execution"
	runservice "polyglot-rosetta/internal/run"
)

type nativeHost struct{ concept runservice.Concept }

func (h nativeHost) Resolve(_ context.Context, id string) (catalog.Concept, error) {
	normalized, err := catalog.NormalizeConceptID(id)
	if err != nil {
		return catalog.Concept{}, err
	}
	if !strings.HasPrefix(h.concept.ID, normalized) {
		return catalog.Concept{}, fmt.Errorf("request ID %q does not match the resolved concept", id)
	}
	tests, err := runservice.LoadTestCases(h.concept)
	return catalog.Concept{Path: h.concept.Reference, TestCases: tests}, err
}

func (h nativeHost) EnsureAvailable(_ context.Context, name string) error {
	lang, err := language.Parse(name)
	if err != nil {
		return err
	}
	return runservice.EnsureAvailable(lang)
}

func (h nativeHost) Execute(ctx context.Context, request runapi.Request, tests []testcase.TestCase) (runapi.Result, error) {
	lang, err := language.Parse(request.Language)
	if err != nil {
		return runapi.Result{}, err
	}
	evaluated, err := engine.EvaluateSuite(ctx, evaluation.Request{Language: lang, SourceCode: request.Source, Mode: request.Mode}, &testcase.TestSuite{Concept: h.concept.Reference, TestCases: tests})
	result := runapi.NewResult()
	result.ID, result.Language, result.Mode = request.ID, request.Language, request.Mode
	if err != nil {
		return result, err
	}
	return resultFromEvaluation(result, evaluated), nil
}
