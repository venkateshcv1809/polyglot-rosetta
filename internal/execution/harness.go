package execution

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"strconv"
	"text/template"
	"time"

	"polyglot-rosetta/internal/api"
	executionapi "polyglot-rosetta/internal/api/execution"
	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/domain/testcase"
)

// These are executable language sources with only input data and the result
// marker substituted. Execution and comparison stay in Go.
//
//go:embed harnesses/*.tmpl
var harnessFiles embed.FS

var harnessTemplates = template.Must(template.ParseFS(harnessFiles, "harnesses/*.tmpl"))

func prepareProgram(lang language.Language, source string, testCases []testcase.TestCase, marker string, timeout time.Duration) (executionapi.Program, error) {
	definition, ok := languagePrograms[lang]
	if !ok {
		return executionapi.Program{}, fmt.Errorf("unsupported language %q", lang)
	}
	harness, err := renderHarness(lang, testCases, marker)
	if err != nil {
		return executionapi.Program{}, err
	}
	files := map[string]string{"main." + lang.FileExtension(): harness, "solution." + lang.FileExtension(): source}
	for name, content := range definition.files {
		files[name] = content
	}

	return executionapi.Program{
		SchemaVersion: api.SchemaVersion, SessionID: marker, Language: string(lang),
		Entry: "main." + lang.FileExtension(), Files: files,
		BuildRequired: definition.build != nil, TimeoutMs: durationMilliseconds(max(timeout, minimumBuildTimeout)),
	}, nil
}

type harnessCase struct {
	Index      int
	Input      string
	TextOutput bool
}

func renderHarness(lang language.Language, testCases []testcase.TestCase, marker string) (string, error) {
	data := struct {
		Marker string
		Inputs string
		Cases  []harnessCase
	}{Marker: strconv.Quote(marker)}
	inputs := make([]map[string]any, len(testCases))
	for i, test := range testCases {
		input, err := json.Marshal(test.Input)
		if err != nil {
			return "", fmt.Errorf("encode input for testcase %q: %w", test.ID, err)
		}
		if _, err := json.Marshal(test.Expected); err != nil {
			return "", fmt.Errorf("encode expected value for testcase %q: %w", test.ID, err)
		}
		_, textOutput := test.Expected.(string)
		entry := harnessCase{Index: i, TextOutput: textOutput}
		switch lang {
		case language.Go:
			entry.Input = strconv.Quote(string(input))
		case language.Rust:
			entry.Input, err = rustInput(input)
			if err != nil {
				return "", err
			}
		case language.Zig:
			entry.Input = "{" + byteList(input) + "}"
		}
		inputs[i] = test.Input
		data.Cases = append(data.Cases, entry)
	}
	payload, err := json.Marshal(inputs)
	if err != nil {
		return "", err
	}
	// A JSON string literal is also valid in the Python and TypeScript templates.
	quoted, _ := json.Marshal(string(payload))
	data.Inputs = string(quoted)
	var rendered bytes.Buffer
	if err := harnessTemplates.ExecuteTemplate(&rendered, string(lang)+".tmpl", data); err != nil {
		return "", fmt.Errorf("render %s harness: %w", lang, err)
	}
	return rendered.String(), nil
}
