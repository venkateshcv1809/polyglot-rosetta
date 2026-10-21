// Package toolchain inspects native language tools available on the host.
package toolchain

import "polyglot-rosetta/internal/domain/language"

// Tool records the outcome of inspecting one required binary.
type Tool struct {
	Binary  string
	Path    string
	Version string
	Error   error
}

// Result records the inspection outcomes for one language.
type Result struct {
	Language language.Language
	Tools    []Tool
}

// Probe inspects every required binary for the selected languages. The lookup
// and version functions are supplied by the caller so probing can be tested
// without depending on the host's installed toolchains.
func Probe(languages []language.Language, lookPath func(string) (string, error), inspect func(string, string) (string, error)) []Result {
	results := make([]Result, 0, len(languages))
	for _, lang := range languages {
		result := Result{Language: lang}
		for _, binary := range RequiredBinaries(lang) {
			tool := Tool{Binary: binary}
			path, err := lookPath(binary)
			if err != nil {
				tool.Error = err
				result.Tools = append(result.Tools, tool)
				continue
			}
			tool.Path = path
			version, err := inspect(binary, path)
			if err != nil {
				tool.Error = err
				result.Tools = append(result.Tools, tool)
				continue
			}
			tool.Version = version
			result.Tools = append(result.Tools, tool)
		}
		results = append(results, result)
	}
	return results
}

// RequiredBinaries lists the native programs needed by a language.
func RequiredBinaries(lang language.Language) []string {
	switch lang {
	case language.Go:
		return []string{"go"}
	case language.Python:
		return []string{"python3"}
	case language.Rust:
		return []string{"rustc"}
	case language.TypeScript:
		return []string{"tsc", "node"}
	case language.Zig:
		return []string{"zig"}
	default:
		return nil
	}
}
