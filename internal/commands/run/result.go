package runcmd

import (
	"time"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/domain/evaluation"
	engine "polyglot-rosetta/internal/execution"
)

func initialResult(opts options) runapi.Result {
	mode := runapi.ModePublic
	if opts.all {
		mode = runapi.ModeAll
	}
	return runapi.Result{
		ID: opts.id, Language: opts.language, Mode: mode,
		Status: evaluation.StatusError, TestResults: []runapi.TestResult{},
	}
}

func resultFromEvaluation(result runapi.Result, evaluated engine.Evaluation) runapi.Result {
	result.DurationMs = evaluated.Duration.Round(time.Millisecond).Milliseconds()
	if evaluated.BuildDuration != nil {
		buildDurationMs := evaluated.BuildDuration.Round(time.Millisecond).Milliseconds()
		result.BuildDurationMs = &buildDurationMs
	}
	result.TotalPassed = evaluated.Report.PassedTests
	result.TotalFailed = evaluated.Report.TotalTests - evaluated.Report.PassedTests
	result.TestResults = make([]runapi.TestResult, 0, len(evaluated.Report.Outcomes))
	for _, outcome := range evaluated.Report.Outcomes {
		result.TestResults = append(result.TestResults, runapi.TestResult{
			ID: outcome.TestCaseID, Status: outcome.Status, Passed: outcome.Passed,
			Hidden: outcome.Hidden, Message: outcome.Message, SolutionTimeNs: outcome.SolutionTimeNs,
		})
	}
	result.Status = evaluation.StatusPassed
	if !evaluated.Report.PassedAll {
		result.Status = evaluation.StatusFailed
	}
	return result
}
