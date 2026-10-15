package scaffoldcmd

import (
	"embed"
	"fmt"
	"io"
	"path/filepath"
	"text/template"

	"polyglot-rosetta/internal/config"
	"polyglot-rosetta/internal/domain/catalog"
	"polyglot-rosetta/internal/scaffold"
)

//go:embed output/*.tmpl
var outputFiles embed.FS

func printCreatedCategory(output io.Writer, opts scaffold.CreateCategoryOptions) error {
	categoryPath := filepath.ToSlash(filepath.Join(opts.ParentPath, opts.ID))
	return renderOutput(output, "category.tmpl", struct {
		Path string
	}{Path: displayPath(categoryPath)})
}

func printCreatedConcept(output io.Writer, opts scaffold.CreateConceptOptions) error {
	conceptPath := filepath.ToSlash(filepath.Join(opts.ParentPath, opts.ID))
	runID, err := catalog.ConceptID(conceptPath)
	if err != nil {
		return fmt.Errorf("derive run ID for %q: %w", conceptPath, err)
	}
	return renderOutput(output, "concept.tmpl", struct {
		Path    string
		ShortID string
		FullID  string
	}{
		Path:    displayPath(conceptPath),
		ShortID: runID[:2+catalog.MinConceptIDHexDigits],
		FullID:  runID,
	})
}

func displayPath(relative string) string {
	return filepath.ToSlash(filepath.Join(config.CatalogDir, filepath.FromSlash(relative)))
}

func renderOutput(output io.Writer, name string, data any) error {
	tmpl, err := template.New("scaffold-output").Option("missingkey=error").ParseFS(outputFiles, "output/*.tmpl")
	if err != nil {
		return fmt.Errorf("parse scaffold output templates: %w", err)
	}
	if err := tmpl.ExecuteTemplate(output, name, data); err != nil {
		return fmt.Errorf("render scaffold %s: %w", name, err)
	}
	return nil
}
