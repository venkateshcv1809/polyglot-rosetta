package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// RequestsJSON reports whether args select JSON output before flags are parsed.
func RequestsJSON(args []string) bool {
	format := ""
	for i, arg := range args {
		if IsHelp(arg) {
			return false
		}
		if arg == "--format" || arg == "-format" {
			if i+1 < len(args) {
				format = args[i+1]
			}
			continue
		}
		if value, found := strings.CutPrefix(arg, "--format="); found {
			format = value
		} else if value, found := strings.CutPrefix(arg, "-format="); found {
			format = value
		}
	}
	return format == "json"
}

// ValidateOutputFormat accepts the public text and JSON presentation modes.
func ValidateOutputFormat(format string) error {
	if format != "text" && format != "json" {
		return fmt.Errorf("unsupported output format %q: use text or json", format)
	}
	return nil
}

// DiscardUsageIfJSON hides flag usage when the caller asked for JSON errors.
func DiscardUsageIfJSON(parser *FlagParser, args []string, output io.Writer) {
	if RequestsJSON(args) {
		parser.SetOutput(io.Discard)
		return
	}
	parser.SetOutput(output)
}

// WriteJSON encodes one result object to output.
func WriteJSON(output io.Writer, result any) error {
	if err := json.NewEncoder(output).Encode(result); err != nil {
		return fmt.Errorf("write JSON result: %w", err)
	}
	return nil
}

// ReportJSON writes a structured result and preserves the failure exit status.
func ReportJSON(output io.Writer, result any, err error) error {
	if writeErr := WriteJSON(output, result); writeErr != nil {
		return writeErr
	}
	return ReportedError(err)
}
