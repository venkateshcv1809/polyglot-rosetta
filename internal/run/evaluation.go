package run

import (
	"context"
	"fmt"
	"os"
	"strings"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/catalog"
	catalogcontent "polyglot-rosetta/internal/catalog/content"
	"polyglot-rosetta/internal/domain/execution"
	"polyglot-rosetta/internal/domain/language"
	engine "polyglot-rosetta/internal/execution"
)

// Evaluate runs a validated request against a previously resolved concept.
func Evaluate(ctx context.Context, request runapi.Request, concept Concept) (engine.Evaluation, error) {
	if err := request.Validate(); err != nil {
		return engine.Evaluation{}, err
	}
	id, err := catalog.NormalizeConceptID(request.ID)
	if err != nil {
		return engine.Evaluation{}, err
	}
	if concept.directory == "" || !strings.HasPrefix(concept.ID, id) {
		return engine.Evaluation{}, fmt.Errorf("request ID %q does not match the resolved concept", request.ID)
	}
	lang, err := language.Parse(request.Language)
	if err != nil {
		return engine.Evaluation{}, err
	}
	if err := RequireToolchain(lang); err != nil {
		return engine.Evaluation{}, err
	}
	suite, err := catalogcontent.LoadTestSuite(os.DirFS(concept.root), concept.Reference)
	if err != nil {
		return engine.Evaluation{}, err
	}
	return engine.EvaluateSuite(ctx, execution.Request{
		Language: lang, SourceCode: request.Source,
		IsSubmit: request.Mode == runapi.ModeAll,
	}, suite)
}
