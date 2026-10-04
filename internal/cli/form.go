package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// PromptField defines one value collected by an interactive form.
type PromptField struct {
	// Key identifies the value in the input and result maps.
	Key string
	// Prompt is the user-facing label shown before reading input.
	Prompt string
	// Required repeats the prompt when the user submits an empty value.
	Required bool
	// DefaultValue is used when the user submits an empty value.
	DefaultValue string
}

// PromptForm collects missing field values from input and writes prompts to
// output. Existing non-empty values are preserved and are not prompted again.
func PromptForm(input io.Reader, output io.Writer, fields []PromptField, existingValues map[string]string) (map[string]string, error) {
	reader := bufio.NewReader(input)
	result := make(map[string]string, len(existingValues)+len(fields))
	for key, value := range existingValues {
		result[key] = value
	}

	for _, field := range fields {
		if value, exists := result[field.Key]; exists && strings.TrimSpace(value) != "" {
			continue
		}

		for {
			if err := writePrompt(output, field); err != nil {
				return nil, fmt.Errorf("write prompt for %q: %w", field.Key, err)
			}

			line, err := reader.ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("read value for %q: %w", field.Key, err)
			}
			if errors.Is(err, io.EOF) && line == "" {
				return nil, fmt.Errorf("read value for %q: %w", field.Key, io.EOF)
			}

			value := strings.TrimSpace(line)
			if value == "" {
				value = field.DefaultValue
			}
			if value == "" && field.Required {
				if _, writeErr := fmt.Fprintln(output, "  Error: this field is required."); writeErr != nil {
					return nil, fmt.Errorf("write validation error for %q: %w", field.Key, writeErr)
				}
				continue
			}

			result[field.Key] = value
			break
		}
	}

	return result, nil
}

func writePrompt(output io.Writer, field PromptField) error {
	if field.DefaultValue != "" {
		_, err := fmt.Fprintf(output, "%s [%s]: ", field.Prompt, field.DefaultValue)
		return err
	}
	_, err := fmt.Fprintf(output, "%s: ", field.Prompt)
	return err
}
