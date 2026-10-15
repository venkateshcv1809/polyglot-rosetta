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
	languages := language.AllSupported()
	if opts.language != "" {
		languages = []language.Language{language.Language(opts.language)}
	}
	result := resultFromToolchain(toolchain.Probe(languages, lookPath, inspect))
	if err := writeResult(output, opts.format, result); err != nil {
		return err
	}
	if result.Status == probeapi.StatusUnavailable {
		if opts.format == "json" {
			return cli.ReportedError(errToolchainsUnavailable)
		}
		return errToolchainsUnavailable
	}
	return nil
}
