package probecmd

import (
	"errors"
	"io"
	"os"
	"os/exec"

	probeapi "polyglot-rosetta/internal/api/probe"
	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/toolchain"
)

var errToolchainsUnavailable = errors.New("one or more required toolchains are unavailable")

func probeCommand(commandPath string, args []string) error {
	return probeCommandWithIO(commandPath, args, os.Stdout, exec.LookPath, toolchain.InspectBinary)
}

func probeCommandWithIO(commandPath string, args []string, output io.Writer, lookPath func(string) (string, error), inspect func(string, string) (string, error)) error {
	opts, err := parseOptionsWithOutput(commandPath, args, output)
	if err != nil {
		if !requestsJSON(args) {
			return err
		}
		result := probeapi.Result{Status: probeapi.StatusError, Toolchains: []probeapi.Toolchain{}, Error: err.Error()}
		if writeErr := writeResult(output, "json", result); writeErr != nil {
			return writeErr
		}
		return cli.ReportedError(err)
	}
	return probeWithInspector(opts, output, lookPath, inspect)
}

func probeWithInspector(opts options, output io.Writer, lookPath func(string) (string, error), inspect func(string, string) (string, error)) error {
	cachePath, err := toolchain.DefaultCachePath()
	if err != nil {
		if opts.clear && opts.format != "json" {
			return cli.CommandFailure("Toolchain cache not cleared", err)
		}
		result := probeapi.Result{Status: probeapi.StatusError, Toolchains: []probeapi.Toolchain{}, Error: err.Error()}
		if writeErr := writeResult(output, opts.format, result); writeErr != nil {
			return writeErr
		}
		if opts.format == "json" {
			return cli.ReportedError(err)
		}
		return cli.CommandFailure("Toolchain cache not updated", err)
	}
	if opts.clear {
		if err := toolchain.ClearCache(cachePath); err != nil {
			if opts.format != "json" {
				return cli.CommandFailure("Toolchain cache not cleared", err)
			}
			result := probeapi.Result{Status: probeapi.StatusError, Toolchains: []probeapi.Toolchain{}, Error: err.Error()}
			if writeErr := writeResult(output, opts.format, result); writeErr != nil {
				return writeErr
			}
			return cli.ReportedError(err)
		}
		return writeResult(output, opts.format, probeapi.Result{Status: probeapi.StatusCleared, Toolchains: []probeapi.Toolchain{}})
	}
	results, refreshErr := toolchain.RefreshCache(cachePath, lookPath, inspect)
	if opts.language != "" {
		selected := make([]toolchain.Result, 0, 1)
		for _, item := range results {
			if item.Language == language.Language(opts.language) {
				selected = append(selected, item)
				break
			}
		}
		results = selected
	}
	result := resultFromToolchain(results)
	if refreshErr != nil {
		result.Status = probeapi.StatusError
		result.Error = refreshErr.Error()
	}
	if err := writeResult(output, opts.format, result); err != nil {
		return err
	}
	if refreshErr != nil {
		if opts.format == "json" {
			return cli.ReportedError(refreshErr)
		}
		return cli.CommandFailure("Toolchain cache not updated", refreshErr)
	}
	if result.Status == probeapi.StatusUnavailable {
		if opts.format == "json" {
			return cli.ReportedError(errToolchainsUnavailable)
		}
		return errToolchainsUnavailable
	}
	return nil
}
