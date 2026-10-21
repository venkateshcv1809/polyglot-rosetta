package listcmd

import (
	"embed"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"text/template"

	"polyglot-rosetta/internal/catalog"
)

//go:embed output/*.tmpl
var outputFiles embed.FS

type displayRow struct {
	Path    string
	ShortID string
	ID      string
	Title   string
}

func writeList(output io.Writer, rows []listedConcept, search string) error {
	data := struct {
		Search string
		Rows   []displayRow
	}{Rows: make([]displayRow, 0, len(rows))}
	if strings.TrimSpace(search) != "" {
		data.Search = search
	}
	for _, item := range rows {
		data.Rows = append(data.Rows, displayRow{
			Path: item.concept.Path, ShortID: item.id[:2+catalog.MinConceptIDHexDigits],
			ID: item.id, Title: item.concept.Title,
		})
	}
	name := "list.tmpl"
	if len(data.Rows) == 0 {
		name = "empty.tmpl"
	}
	tmpl, err := template.New("list-output").Option("missingkey=error").ParseFS(outputFiles, "output/*.tmpl")
	if err != nil {
		return fmt.Errorf("parse list output templates: %w", err)
	}
	writer := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)
	if err := tmpl.ExecuteTemplate(writer, name, data); err != nil {
		return fmt.Errorf("render list %s: %w", name, err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush list output: %w", err)
	}
	return nil
}
