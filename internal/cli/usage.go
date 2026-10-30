package cli

import (
	"fmt"
	"io"
)

// Option describes one command-line option in help output.
type Option struct {
	// Name is the long option name without leading dashes.
	Name string
	// Value is the placeholder displayed for options that accept a value.
	Value string
	// Description explains the option's effect.
	Description string
	// Required marks the option as required in help; callers enforce validation.
	Required bool
	// Default overrides the default value displayed in help when its semantic
	// value differs from the parser's literal default.
	Default string
}

// PrintUsage writes consistently formatted help for a leaf command.
func PrintUsage(w io.Writer, commandPath string, forms []string, options []Option) {
	fmt.Fprintln(w, "Usage:")
	for _, form := range forms {
		if form == "" {
			fmt.Fprintf(w, "  %s\n", commandPath)
			continue
		}
		fmt.Fprintf(w, "  %s %s\n", commandPath, form)
	}

	if len(options) == 0 {
		return
	}

	fmt.Fprintln(w, "\nOptions:")
	width := 0
	labels := make([]string, len(options))
	for i, option := range options {
		labels[i] = "--" + option.Name
		if option.Value != "" {
			labels[i] += "=<" + option.Value + ">"
		}
		if len(labels[i]) > width {
			width = len(labels[i])
		}
	}

	for i, option := range options {
		description := option.Description
		if option.Required {
			description += " (required)"
		} else if option.Default != "" {
			description += fmt.Sprintf(" (default: %s)", option.Default)
		}
		fmt.Fprintf(w, "  %-*s  %s\n", width, labels[i], description)
	}
}
