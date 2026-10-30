package run

import (
	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/domain/language"
)

// Prepare resolves a registered concept and chooses explicit or catalog source.
// It returns the resolved concept even when source loading or validation fails.
func Prepare(id, languageName string, mode runapi.Mode, source *string) (runapi.Request, Concept, error) {
	concept, err := ResolveConcept(id)
	if err != nil {
		return runapi.Request{}, Concept{}, err
	}
	lang, err := language.Parse(languageName)
	if err != nil {
		return runapi.Request{}, concept, err
	}
	code := ""
	if source == nil {
		code, err = LoadSolution(concept, lang)
		if err != nil {
			return runapi.Request{}, concept, err
		}
	} else {
		code = *source
	}
	request := runapi.Request{ID: concept.ID, Language: languageName, Source: code, Mode: mode}
	if err := request.Validate(); err != nil {
		return runapi.Request{}, concept, err
	}
	return request, concept, nil
}
