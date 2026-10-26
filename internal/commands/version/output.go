package versioncmd

import (
	"embed"
	"fmt"
	"io"
	"text/template"

	versionapi "polyglot-rosetta/internal/api/version"
	"polyglot-rosetta/internal/cli"
)

//go:embed output/*.tmpl
var outputFiles embed.FS

func writeResult(output io.Writer, format string, result versionapi.Result) error {
	if format == "json" {
		return cli.WriteJSON(output, result)
	}
	tmpl, err := template.New("version.tmpl").Option("missingkey=error").ParseFS(outputFiles, "output/*.tmpl")
	if err != nil {
		return fmt.Errorf("parse version output template: %w", err)
	}
	if err := tmpl.ExecuteTemplate(output, "version.tmpl", result); err != nil {
		return fmt.Errorf("render version output: %w", err)
	}
	return nil
}
