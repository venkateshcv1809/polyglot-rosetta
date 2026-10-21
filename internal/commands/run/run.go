package runcmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	iofs "io/fs"
	"os"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/domain/evaluation"
	"polyglot-rosetta/internal/domain/language"
	runservice "polyglot-rosetta/internal/run"
)

var errTestsFailed = errors.New("one or more test cases failed")

type preparedRun struct {
	result  runapi.Result
	request runapi.Request
	concept runservice.Concept
}

func runCommand(commandPath string, args []string) error {
	return runCommandWithIO(context.Background(), commandPath, args, os.Stdin, os.Stdout)
}

func runCommandWithIO(ctx context.Context, commandPath string, args []string, stdin io.Reader, stdout io.Writer) error {
	opts, err := parseOptionsWithOutput(commandPath, args, stdout)
	if err != nil {
		if !requestsJSON(args) {
			return err
		}
		result := initialResult(opts)
		result.Error = err.Error()
		if writeErr := writeResult(stdout, "json", result); writeErr != nil {
			return writeErr
		}
		return cli.ReportedError(err)
	}
	return runWithIO(ctx, opts, stdin, stdout)
}

func runWithIO(ctx context.Context, opts options, stdin io.Reader, stdout io.Writer) error {
	prepared, runErr := prepareRun(opts, stdin)
	if runErr == nil {
		runErr = runservice.RequireToolchain(language.Language(opts.language))
	}
	if runErr == nil {
		evaluation, err := runservice.Evaluate(ctx, prepared.request, prepared.concept)
		runErr = err
		if err == nil {
			prepared.result = resultFromEvaluation(prepared.result, evaluation)
		}
	}
	if runErr != nil {
		prepared.result.Error = runErr.Error()
	}
	if err := writeResult(stdout, opts.format, prepared.result); err != nil {
		return err
	}
	if runErr != nil {
		if opts.format == "json" {
			return cli.ReportedError(runErr)
		}
		return runErr
	}
	if prepared.result.Status == evaluation.StatusFailed {
		if opts.format == "json" {
			return cli.ReportedError(errTestsFailed)
		}
		return errTestsFailed
	}
	return nil
}

func prepareRun(opts options, stdin io.Reader) (preparedRun, error) {
	prepared := preparedRun{result: initialResult(opts)}
	source, err := sourceCode(opts, stdin)
	if err != nil {
		return prepared, err
	}
	request, concept, err := runservice.Prepare(opts.id, opts.language, prepared.result.Mode, source)
	if concept.ID != "" {
		prepared.concept = concept
		prepared.result.Concept = concept.Reference
		prepared.result.ID = concept.ID
	}
	if err != nil {
		if errors.Is(err, runservice.ErrCatalogUnavailable) {
			return prepared, fmt.Errorf("%w: run prosetta catalog --url=<directory-url> or set PROSETTA_CATALOG", err)
		}
		if source == nil && errors.Is(err, iofs.ErrNotExist) {
			return prepared, fmt.Errorf("%w (use --stdin or --code to supply source)", err)
		}
		return prepared, err
	}
	switch {
	case opts.stdin:
		prepared.result.InputSource = "stdin"
	case opts.code != "":
		prepared.result.InputSource = "inline"
	default:
		prepared.result.InputSource = "catalog"
	}
	prepared.request = request
	return prepared, nil
}

func sourceCode(opts options, stdin io.Reader) (*string, error) {
	switch {
	case opts.stdin:
		data, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read from standard input: %w", err)
		}
		value := string(data)
		return &value, nil
	case opts.code != "":
		return &opts.code, nil
	default:
		return nil, nil
	}
}
