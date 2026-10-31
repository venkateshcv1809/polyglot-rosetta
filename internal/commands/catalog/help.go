//go:build js && wasm

package catalogcmd

import (
	"polyglot-rosetta/internal/wasm"
)

// wasmHelp describes the public WASM operation separately from its handler.
func wasmHelp() wasm.MethodHelp {
	return wasm.MethodHelp{
		Description: "Read or activate catalog metadata and category archives supplied by the host.",
		Signatures:  []string{"catalog()", "catalog(index)", "catalog(conceptID)", "catalog(index, templates, archives)"},
		Parameters: []wasm.Option{
			{Name: "indexOrConceptID", Type: "Uint8Array|string", Description: "Index bytes to activate metadata, or a full concept ID/unique prefix to resolve. Omit to read the current catalog."},
			{Name: "templates", Type: "Uint8Array", Description: "Template archive bytes; required together with archives when activating categories."},
			{Name: "archives", Type: "object", Description: "Category filename to Uint8Array mapping, e.g. {'fundamentals.tar.gz': bytes}; required together with templates."},
		},
		Returns: "CatalogResult {schema_version, status, catalog?, concept_id?, error?}",
		Notes:   []string{"Byte arguments must be non-empty and at most 128 MiB each. Hosts control downloads and verify asset hashes.", "Index-only activation exposes metadata. Category activation supplies exercise content. Failed activation preserves the previous snapshot."},
	}
}
