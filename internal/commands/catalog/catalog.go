package catalogcmd

import (
	"context"
	"io"
	"os"

	catalogapi "polyglot-rosetta/internal/api/catalog"
	catalogsync "polyglot-rosetta/internal/catalog/sync"
	"polyglot-rosetta/internal/cli"
)

type syncCatalog func(context.Context, string, string) (string, error)

func catalogCommand(commandPath string, args []string) error {
	return catalogCommandWithIO(context.Background(), commandPath, args, os.Stdout, catalogsync.Sync)
}

func catalogCommandWithIO(ctx context.Context, commandPath string, args []string, output io.Writer, sync syncCatalog) error {
	opts, err := parseOptionsWithOutput(commandPath, args, output)
	if err != nil {
		if !requestsJSON(args) {
			return err
		}
		result := catalogapi.Result{Status: catalogapi.StatusError, Error: err.Error()}
		if writeErr := writeResult(output, "json", result); writeErr != nil {
			return writeErr
		}
		return cli.ReportedError(err)
	}
	return catalogWithSync(ctx, opts, output, sync)
}

func catalogWithSync(ctx context.Context, opts options, output io.Writer, sync syncCatalog) error {
	root, err := sync(ctx, opts.url, opts.cache)
	if err != nil {
		if opts.format != "json" {
			return cli.CommandFailure("Catalog not ready", err)
		}
		if writeErr := writeResult(output, opts.format, catalogapi.Result{Status: catalogapi.StatusError, Error: err.Error()}); writeErr != nil {
			return writeErr
		}
		return cli.ReportedError(err)
	}
	return writeResult(output, opts.format, catalogapi.Result{Status: catalogapi.StatusOK, Root: root})
}
