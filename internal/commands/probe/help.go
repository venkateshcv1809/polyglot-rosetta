//go:build js && wasm

package probecmd

import (
	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/wasm"
)

// wasmHelp describes the public WASM operation separately from its handler.
func wasmHelp() wasm.MethodHelp {
	languages := make([]string, 0)
	for _, lang := range language.AllSupported() {
		languages = append(languages, string(lang))
	}
	return wasm.MethodHelp{
		Description: "Refresh and report browser toolchain readiness, or clear only its saved snapshot.",
		Async:       true,
		Signatures:  []string{"probe()", "probe(language)", "probe({language?, clear?})"},
		Parameters: []wasm.Option{{Name: "request", Type: "string|object", Description: "A language filter or probe options; omit to inspect every language.", Properties: []wasm.Option{
			{Name: "language", Type: "string", Description: "Language filter; empty means all languages. Supported languages must also be supplied by the host.", Values: languages},
			{Name: "clear", Type: "boolean", Default: false, Description: "Delete only the readiness snapshot without inspecting or deleting compiler assets."},
		}}},
		Callbacks: []string{"refreshCache"},
		Returns:   "Promise<ProbeResult> {schema_version, status: ok|unavailable|cleared|error, toolchains, error?}",
		Notes:     []string{"clear: true uses clearCache instead of refreshCache and cannot be combined with a language filter."},
	}
}
