// Package execution evaluates submitted source code in temporary workspaces
// using the host's native language toolchains. It limits time and captured
// output, but it is not a security boundary for hostile source code.
package execution

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"polyglot-rosetta/internal/api"
	executionapi "polyglot-rosetta/internal/api/execution"
	domainevaluation "polyglot-rosetta/internal/domain/evaluation"
	"polyglot-rosetta/internal/domain/testcase"
)

const defaultExecutionTimeout = 10 * time.Second

// Evaluation contains a completed test report and captured process output.
type Evaluation struct {
	Report        domainevaluation.EvaluationReport
	Duration      time.Duration
	BuildDuration *time.Duration
	Stdout        string
	Stderr        string
}

// EvaluateSuite evaluates a request against an already loaded test suite.
func EvaluateSuite(ctx context.Context, request domainevaluation.Request, suite *testcase.TestSuite) (Evaluation, error) {
	return EvaluateSuiteWithTimeout(ctx, request, suite, defaultExecutionTimeout)
}

// EvaluateSuiteWithTimeout evaluates an in-memory suite with a caller-supplied timeout.
func EvaluateSuiteWithTimeout(ctx context.Context, request domainevaluation.Request, suite *testcase.TestSuite, timeout time.Duration) (evaluation Evaluation, err error) {
	return evaluateSuite(ctx, request, suite, timeout, NewNativeHost())
}

// EvaluateSuiteWithHost runs the same evaluation pipeline on native or browser
// hosts. Only compilation, launching one case, and session cleanup are delegated.
func EvaluateSuiteWithHost(ctx context.Context, request domainevaluation.Request, suite *testcase.TestSuite, host Host) (Evaluation, error) {
	return evaluateSuite(ctx, request, suite, defaultExecutionTimeout, host)
}

func evaluateSuite(ctx context.Context, request domainevaluation.Request, suite *testcase.TestSuite, timeout time.Duration, host Host) (evaluation Evaluation, err error) {
	started := time.Now()
	defer func() { evaluation.Duration = time.Since(started) }()
	if timeout <= 0 {
		return Evaluation{}, fmt.Errorf("execution timeout must be positive")
	}
	if suite == nil {
		return Evaluation{}, fmt.Errorf("test suite cannot be nil")
	}
	testCases := suite.GetTestCases(request.Mode.IncludeHidden())
	if len(testCases) == 0 {
		return Evaluation{}, fmt.Errorf("concept %q has no selected test cases", suite.Concept)
	}

	marker, err := newResultMarker()
	if err != nil {
		return Evaluation{}, err
	}
	program, err := prepareProgram(request.Language, request.SourceCode, testCases, marker, timeout)
	if err != nil {
		return Evaluation{}, fmt.Errorf("prepare %s evaluation: %w", request.Language, err)
	}
	// Dispose even if the build partially allocates resources and then fails.
	defer func() {
		if cleanupErr := host.DisposeProgram(context.WithoutCancel(ctx), program.SessionID); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("dispose evaluation program: %w", cleanupErr))
		}
	}()
	built, err := host.BuildProgram(ctx, program)
	if err != nil {
		return evaluation, err
	}
	if err := validateBuild(program, built); err != nil {
		return evaluation, err
	}
	if built.BuildDurationMs != nil {
		duration := time.Duration(*built.BuildDurationMs) * time.Millisecond
		evaluation.BuildDuration = &duration
	}
	if built.Error != "" {
		return evaluation, errors.New(built.Error)
	}

	outcomes := make([]domainevaluation.TestOutcome, 0, len(testCases))
	stdout := newTailBuffer(defaultCommandOutputLimit)
	stderr := newTailBuffer(defaultCommandOutputLimit)
	for index, testCase := range testCases {
		caseRequest := executionapi.CaseRequest{
			SchemaVersion: api.SchemaVersion, SessionID: program.SessionID, Index: index,
			Args: []string{strconv.Itoa(index)}, TimeoutMs: max(1, durationMilliseconds(timeout)), MaxOutputBytes: defaultCommandOutputLimit,
		}
		raw, err := host.RunCase(ctx, caseRequest)
		if err != nil {
			return evaluation, err
		}
		result, err := processResult(caseRequest, raw)
		if err != nil {
			return evaluation, err
		}
		if err := ctx.Err(); err != nil {
			return evaluation, err
		}
		_, _ = stdout.Write([]byte(result.Stdout))
		_, _ = stderr.Write([]byte(result.Stderr))
		outcomes = append(outcomes, caseOutcome(testCase, index, result, marker))
	}

	selectedSuite := &testcase.TestSuite{Concept: suite.Concept, TestCases: testCases}
	evaluation.Report = domainevaluation.AggregateResults(selectedSuite, outcomes)
	evaluation.Stdout, evaluation.Stderr = stdout.String(), stderr.String()
	return evaluation, nil
}

func caseOutcome(testCase testcase.TestCase, index int, result RunResult, marker string) domainevaluation.TestOutcome {
	outcome := domainevaluation.TestOutcome{TestCaseID: testCase.ID, Hidden: testCase.Hidden}
	switch {
	case result.TimedOut:
		outcome.Status = domainevaluation.TestTimeout
		outcome.Message = "solution exceeded the time limit"
	case result.Error != nil:
		outcome.Status = domainevaluation.TestCrashed
		if result.ExitCode >= 0 {
			outcome.Message = fmt.Sprintf("solution exited with code %d", result.ExitCode)
		} else {
			outcome.Message = "solution process terminated"
		}
	default:
		parsed, err := parseCaseOutcome(index, result.Stdout+"\n"+result.Stderr, marker)
		if err != nil {
			outcome.Status = domainevaluation.TestCrashed
			outcome.Message = "solution exited without a valid result"
		} else {
			outcome.Passed = matchesExpected(parsed, testCase.Expected)
			outcome.SolutionTimeNs = parsed.SolutionTimeNs
			if outcome.Passed {
				outcome.Status = domainevaluation.TestPassed
			} else {
				outcome.Status = domainevaluation.TestWrongAnswer
				outcome.Message = "output did not match the expected value"
			}
		}
	}
	return outcome
}
