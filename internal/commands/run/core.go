package runcmd

import (
	"context"
	"fmt"
	"strings"

	"polyglot-rosetta/internal/api"
	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/domain/evaluation"
	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/domain/testcase"
)

// Host supplies concept data, cached readiness, and an execution backend.
type Host interface {
	Resolve(context.Context, string) (catalog.Concept, error)
	EnsureAvailable(context.Context, string) error
	Execute(context.Context, runapi.Request, []testcase.TestCase) (runapi.Result, error)
}

// Run owns request semantics and validates results independently of the host.
func Run(ctx context.Context, request runapi.Request, host Host) (runapi.Result, error) {
	result := runapi.NewResult()
	request.ID = strings.TrimSpace(request.ID)
	if request.Mode == "" {
		request.Mode = evaluation.ModePublic
	}
	result.ID, result.Language, result.Mode, result.InputSource = request.ID, request.Language, request.Mode, runapi.InputSourceInline
	fail := func(err error) (runapi.Result, error) { result.Error = err.Error(); return result, err }
	lang, err := language.Parse(request.Language)
	if err != nil {
		return fail(fmt.Errorf("unsupported language %q", request.Language))
	}
	request.Language, result.Language = string(lang), string(lang)
	if err := request.Validate(); err != nil {
		return fail(err)
	}
	concept, err := host.Resolve(ctx, request.ID)
	if err != nil {
		return fail(err)
	}
	request.ID, err = catalog.ConceptID(concept.Path)
	if err != nil {
		return fail(err)
	}
	result.ID, result.Concept = request.ID, concept.Path
	suite := testcase.TestSuite{TestCases: concept.TestCases}
	tests := suite.GetTestCases(request.Mode.IncludeHidden())
	if len(tests) == 0 {
		return fail(fmt.Errorf("concept %q has no selected test cases", request.ID))
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if err := host.EnsureAvailable(ctx, request.Language); err != nil {
		return fail(err)
	}
	evaluated, err := host.Execute(ctx, request, tests)
	if err != nil {
		return fail(err)
	}
	if err := validateResult(evaluated, request, tests); err != nil {
		return fail(err)
	}
	evaluated.Concept, evaluated.InputSource = concept.Path, runapi.InputSourceInline
	for i := range evaluated.TestResults {
		evaluated.TestResults[i].Hidden = tests[i].Hidden
	}
	return evaluated, nil
}

func validateResult(result runapi.Result, request runapi.Request, tests []testcase.TestCase) error {
	invalid := func(reason string) error { return fmt.Errorf("invalid execution result: %s", reason) }
	if result.SchemaVersion != api.SchemaVersion {
		return invalid(fmt.Sprintf("schema_version must be %d", api.SchemaVersion))
	}
	if result.ID != request.ID || result.Language != request.Language || result.Mode != request.Mode {
		return invalid("request identity does not match")
	}
	if result.DurationMs < 0 || (result.BuildDurationMs != nil && *result.BuildDurationMs < 0) {
		return invalid("durations cannot be negative")
	}
	if len(result.TestResults) != len(tests) {
		return invalid("selected test cases do not match")
	}
	passed, timedOut, crashed := 0, false, false
	for i, item := range result.TestResults {
		if item.ID != tests[i].ID || item.SolutionTimeNs < 0 {
			return invalid("test case identity or timing is invalid")
		}
		switch item.Status {
		case evaluation.TestPassed:
			if !item.Passed {
				return invalid("passed case has passed=false")
			}
			passed++
		case evaluation.TestWrongAnswer, evaluation.TestCrashed, evaluation.TestTimeout:
			if item.Passed {
				return invalid("failed case has passed=true")
			}
			timedOut = timedOut || item.Status == evaluation.TestTimeout
			crashed = crashed || item.Status == evaluation.TestCrashed
		default:
			return invalid("unknown test case status")
		}
	}
	if result.TotalPassed != passed || result.TotalFailed != len(tests)-passed {
		return invalid("totals do not match test cases")
	}
	switch result.Status {
	case evaluation.StatusPassed:
		if passed != len(tests) || result.Error != "" {
			return invalid("inconsistent passed status")
		}
	case evaluation.StatusFailed:
		if passed == len(tests) || result.Error != "" {
			return invalid("inconsistent failed status")
		}
	case evaluation.StatusTimeout:
		if !timedOut {
			return invalid("timeout status has no timed-out case")
		}
	case evaluation.StatusError:
		if !crashed && result.Error == "" {
			return invalid("error status has no error details")
		}
	default:
		return invalid("unknown overall status")
	}
	return nil
}
