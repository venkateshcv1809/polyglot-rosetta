package versioncmd

import (
	"io"
	"os"

	"polyglot-rosetta/internal/cli"
)

func versionCommand(commandPath string, args []string) error {
	return versionCommandWithIO(commandPath, args, os.Stdout)
}

func versionCommandWithIO(commandPath string, args []string, output io.Writer) error {
	opts, err := parseOptions(commandPath, args, output)
	if err != nil {
		if !cli.RequestsJSON(args) {
			return err
		}
		return cli.ReportJSON(output, errorResult(err), err)
	}
	return writeResult(output, opts.format, Version())
}
