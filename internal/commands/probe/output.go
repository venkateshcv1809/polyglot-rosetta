package probecmd

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	probeapi "polyglot-rosetta/internal/api/probe"
)

//go:embed output/*.tmpl
var outputFiles embed.FS

type displayRow struct {
	Status   string
	Language string
	Details  string
}

func writeResult(output io.Writer, format string, result probeapi.Result) error {
	if format == "json" {
		if err := json.NewEncoder(output).Encode(result); err != nil {
			return fmt.Errorf("write JSON result: %w", err)
		}
		return nil
	}
	return writeTextResult(output, result)
}

func writeTextResult(output io.Writer, result probeapi.Result) error {
	rows := make([]displayRow, 0, len(result.Toolchains))
	for _, entry := range result.Toolchains {
		row := displayRow{Status: "OK", Language: entry.Language}
		var paths, missing []string
		for _, tool := range entry.Tools {
			if tool.Available {
				paths = append(paths, fmt.Sprintf("%s [%s]", tool.Path, tool.Version))
			} else if tool.Path == "" {
				missing = append(missing, fmt.Sprintf("%q", tool.Binary))
			} else {
				missing = append(missing, fmt.Sprintf("%q", tool.Binary+" (unusable)"))
			}
		}
		if !entry.Available {
			row.Status = "FAIL"
			row.Details = "missing " + strings.Join(missing, ", ")
		} else {
			row.Details = strings.Join(paths, ", ")
		}
		rows = append(rows, row)
	}
	tmpl, err := template.New("probe.tmpl").Option("missingkey=error").ParseFS(outputFiles, "output/*.tmpl")
	if err != nil {
		return fmt.Errorf("parse probe output template: %w", err)
	}
	if err := tmpl.ExecuteTemplate(output, "probe.tmpl", struct{ Rows []displayRow }{Rows: rows}); err != nil {
		return fmt.Errorf("render probe output: %w", err)
	}
	return nil
}
