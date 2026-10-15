package bundlecmd

import (
	"embed"
	"fmt"
	"io"
	"text/template"
)

//go:embed output/*.tmpl
var outputFiles embed.FS

func printCreatedBundle(output io.Writer, path string) error {
	tmpl, err := template.New("bundle.tmpl").Option("missingkey=error").ParseFS(outputFiles, "output/*.tmpl")
	if err != nil {
		return fmt.Errorf("parse bundle output template: %w", err)
	}
	if err := tmpl.ExecuteTemplate(output, "bundle.tmpl", struct{ Path string }{Path: path}); err != nil {
		return fmt.Errorf("render bundle output: %w", err)
	}
	return nil
}
