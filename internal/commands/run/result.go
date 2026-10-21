package runcmd

import (
	"time"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/domain/execution"
	engine "polyglot-rosetta/internal/execution"
)

func initialResult(opts options) runapi.Result {
	mode := runapi.ModePublic
	if opts.all {
		mode = runapi.ModeAll
	}
	return runapi.Result{
		ID: opts.id, Language: opts.language, Mode: mode,
		Status: execution.StatusError, TestResults: []runapi.TestResult{},
	}
}

func resultFromEvaluation(result runapi.Result, evaluation engine.Evaluation) runapi.Result {
	result.DurationMs = evaluation.Duration.Round(time.Millisecond).Milliseconds()
	if evaluation.BuildDuration != nil {
		buildDurationMs := evaluation.BuildDuration.Round(time.Millisecond).Milliseconds()
		result.BuildDurationMs = &buildDurationMs
	}
	result.TotalPassed = evaluation.Report.PassedTests
	result.TotalFailed = evaluation.Report.TotalTests - evaluation.Report.PassedTests
	result.TestResults = make([]runapi.TestResult, 0, len(evaluation.Report.Outcomes))
	for _, outcome := range evaluation.Report.Outcomes {
		result.TestResults = append(result.TestResults, runapi.TestResult{
			ID: outcome.TestCaseID, Status: outcome.Status, Passed: outcome.Passed,
			Hidden: outcome.Hidden, Message: outcome.Message, SolutionTimeNs: outcome.SolutionTimeNs,
		})
	}
	result.Status = execution.StatusPassed
	if !evaluation.Report.PassedAll {
		result.Status = execution.StatusFailed
	}
	return result
}
