package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptField defines the configuration for a single CLI input prompt.
type PromptField struct {
	Key          string
	Prompt       string
	Required     bool
	DefaultValue string
}

// PromptForm dynamically collects input based on the field specification.
func PromptForm(fields []PromptField, existingValues map[string]string) map[string]string {
	reader := bufio.NewReader(os.Stdin)
	result := make(map[string]string)

	for k, v := range existingValues {
		result[k] = v
	}

	for _, field := range fields {
		if val, exists := result[field.Key]; exists && strings.TrimSpace(val) != "" {
			continue
		}

		for {
			promptText := field.Prompt
			if field.DefaultValue != "" {
				promptText = fmt.Sprintf("%s [%s]: ", field.Prompt, field.DefaultValue)
			} else {
				promptText = fmt.Sprintf("%s: ", field.Prompt)
			}

			fmt.Print(promptText)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input == "" && field.DefaultValue != "" {
				input = field.DefaultValue
			}

			if input == "" && field.Required {
				fmt.Println("  ❌ Error: This field is required.")
				continue
			}

			result[field.Key] = input
			break
		}
	}

	return result
}
