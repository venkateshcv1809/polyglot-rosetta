package probecmd

import (
	"errors"
	"io"
	"os"
	"os/exec"

	probeapi "polyglot-rosetta/internal/api/probe"
	"polyglot-rosetta/internal/cli"
	"polyglot-rosetta/internal/toolchain"
)

var errToolchainsUnavailable = errors.New("one or more required toolchains are unavailable")

func probeCommand(commandPath string, args []string) error {
	return probeCommandWithIO(commandPath, args, os.Stdout, exec.LookPath, toolchain.InspectBinary)
}

func probeCommandWithIO(commandPath string, args []string, output io.Writer, lookPath func(string) (string, error), inspect func(string, string) (string, error)) error {
	opts, err := parseOptionsWithOutput(commandPath, args, output)
	if err != nil {
		if !cli.RequestsJSON(args) {
			return err
		}
		return cli.ReportJSON(output, errorResult(err), err)
	}
	return probeWithInspector(opts, output, lookPath, inspect)
}

func probeWithInspector(opts options, output io.Writer, lookPath func(string) (string, error), inspect func(string, string) (string, error)) error {
	cachePath, err := toolchain.DefaultCachePath()
	if err != nil {
		if opts.clear && opts.format != "json" {
			return cli.CommandFailure("Toolchain cache not cleared", err)
		}
		if writeErr := writeResult(output, opts.format, errorResult(err)); writeErr != nil {
			return writeErr
		}
		if opts.format == "json" {
			return cli.ReportedError(err)
		}
		return cli.CommandFailure("Toolchain cache not updated", err)
	}
	if opts.clear {
		result := ClearCache(NativeHost{CachePath: cachePath})
		if result.Status == probeapi.StatusError {
			err := errors.New(result.Error)
			if opts.format != "json" {
				return cli.CommandFailure("Toolchain cache not cleared", err)
			}
			return cli.ReportJSON(output, result, err)
		}
		return writeResult(output, opts.format, result)
	}
	result := Probe(opts.language, NativeHost{CachePath: cachePath, LookPath: lookPath, Inspect: inspect})
	if err := writeResult(output, opts.format, result); err != nil {
		return err
	}
	if result.Status == probeapi.StatusError {
		refreshErr := errors.New(result.Error)
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
