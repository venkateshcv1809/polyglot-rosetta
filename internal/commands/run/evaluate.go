package runcmd

import (
	"context"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/domain/evaluation"
	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/domain/testcase"
	engine "polyglot-rosetta/internal/execution"
)

// Both adapters use the same evaluator and public-result conversion.
func executeEvaluation(ctx context.Context, request runapi.Request, tests []testcase.TestCase, host engine.Host) (runapi.Result, error) {
	result := runapi.NewResult()
	result.ID, result.Language, result.Mode = request.ID, request.Language, request.Mode
	lang, err := language.Parse(request.Language)
	if err != nil {
		return result, err
	}
	evaluated, err := engine.EvaluateSuiteWithHost(ctx, evaluation.Request{
		Language: lang, SourceCode: request.Source, Mode: request.Mode,
	}, &testcase.TestSuite{TestCases: tests}, host)
	result = resultFromEvaluation(result, evaluated)
	if err != nil {
		result.Status, result.Error = evaluation.StatusError, err.Error()
	}
	return result, err
}
