package catalogcmd

import (
	"embed"
	"fmt"
	"io"
	"text/template"

	catalogapi "polyglot-rosetta/internal/api/catalog"
	"polyglot-rosetta/internal/cli"
)

//go:embed output/*.tmpl
var outputFiles embed.FS

func writeResult(output io.Writer, format string, result catalogapi.Result) error {
	if format == "json" {
		return cli.WriteJSON(output, result)
	}
	tmpl, err := template.New("catalog.tmpl").Option("missingkey=error").ParseFS(outputFiles, "output/*.tmpl")
	if err != nil {
		return fmt.Errorf("parse catalog output template: %w", err)
	}
	if err := tmpl.ExecuteTemplate(output, "catalog.tmpl", result); err != nil {
		return fmt.Errorf("render catalog output: %w", err)
	}
	return nil
}
