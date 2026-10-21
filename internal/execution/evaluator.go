// Package execution evaluates submitted source code in temporary workspaces
// using the host's native language toolchains. It limits time and captured
// output, but it is not a security boundary for hostile source code.
package execution

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	domainevaluation "polyglot-rosetta/internal/domain/evaluation"
	"polyglot-rosetta/internal/domain/testcase"
)

const (
	defaultExecutionTimeout = 10 * time.Second
	minimumBuildTimeout     = 60 * time.Second
)

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
	started := time.Now()
	if timeout <= 0 {
		return Evaluation{}, fmt.Errorf("execution timeout must be positive")
	}
	if suite == nil {
		return Evaluation{}, fmt.Errorf("test suite cannot be nil")
	}
	testCases := suite.GetTestCases(request.IsSubmit)
	if len(testCases) == 0 {
		return Evaluation{}, fmt.Errorf("concept %q has no selected test cases", suite.Concept)
	}

	workspace, err := NewWorkspace()
	if err != nil {
		return Evaluation{}, err
	}
	defer func() {
		if cleanupErr := workspace.Cleanup(); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("clean evaluation workspace: %w", cleanupErr))
		}
	}()

	marker, err := newResultMarker()
	if err != nil {
		return Evaluation{}, err
	}
	program, err := prepareProgram(workspace, request.Language, request.SourceCode, testCases, marker)
	if err != nil {
		return Evaluation{}, fmt.Errorf("prepare %s evaluation: %w", request.Language, err)
	}
	var buildDuration *time.Duration
	if len(program.build) > 0 {
		buildDuration = new(time.Duration)
	}
	for _, command := range program.build {
		buildStarted := time.Now()
		result := RunCommand(ctx, max(timeout, minimumBuildTimeout), command)
		*buildDuration += time.Since(buildStarted)
		if result.Error != nil {
			return Evaluation{}, commandError("build", command, result)
		}
	}

	outcomes := make([]domainevaluation.TestOutcome, 0, len(testCases))
	stdout := newTailBuffer(defaultCommandOutputLimit)
	stderr := newTailBuffer(defaultCommandOutputLimit)
	for index, testCase := range testCases {
		command := program.run
		command.Args = append(append([]string(nil), program.run.Args...), strconv.Itoa(index))
		result := RunCommand(ctx, timeout, command)
		if err := ctx.Err(); err != nil {
			return Evaluation{}, err
		}
		_, _ = stdout.Write([]byte(result.Stdout))
		_, _ = stderr.Write([]byte(result.Stderr))
		outcomes = append(outcomes, caseOutcome(testCase, index, result, marker))
	}

	selectedSuite := &testcase.TestSuite{Concept: suite.Concept, TestCases: testCases}
	return Evaluation{
		Report:        domainevaluation.AggregateResults(selectedSuite, outcomes),
		Duration:      time.Since(started),
		BuildDuration: buildDuration,
		Stdout:        stdout.String(),
		Stderr:        stderr.String(),
	}, nil
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
			outcome.Passed = parsed.Passed
			outcome.SolutionTimeNs = parsed.SolutionTimeNs
			if parsed.Passed {
				outcome.Status = domainevaluation.TestPassed
			} else {
				outcome.Status = domainevaluation.TestWrongAnswer
				outcome.Message = "output did not match the expected value"
			}
		}
	}
	return outcome
}

func commandError(stage string, command Command, result RunResult) error {
	if result.TimedOut {
		return fmt.Errorf("%s command %q timed out", stage, command.Name)
	}
	details := strings.TrimSpace(result.Stderr)
	if details == "" {
		details = strings.TrimSpace(result.Stdout)
	}
	if details == "" {
		return fmt.Errorf("%s command %q failed: %w", stage, command.Name, result.Error)
	}
	if result.StdoutTruncated || result.StderrTruncated {
		details = "[output truncated; showing tail]\n" + details
	}
	return fmt.Errorf("%s command %q failed: %w\n%s", stage, command.Name, result.Error, details)
}
