package catalogcmd

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"text/template"

	catalogapi "polyglot-rosetta/internal/api/catalog"
)

//go:embed output/*.tmpl
var outputFiles embed.FS

func writeResult(output io.Writer, format string, result catalogapi.Result) error {
	if format == "json" {
		if err := json.NewEncoder(output).Encode(result); err != nil {
			return fmt.Errorf("write JSON result: %w", err)
		}
		return nil
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
