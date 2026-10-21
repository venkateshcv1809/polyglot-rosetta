package toolchain

import (
	"errors"
	"slices"
	"testing"

	"polyglot-rosetta/internal/domain/language"
)

func TestRequiredBinaries(t *testing.T) {
	tests := map[language.Language][]string{
		language.Go:         {"go"},
		language.Python:     {"python3"},
		language.Rust:       {"rustc"},
		language.TypeScript: {"tsc", "node"},
		language.Zig:        {"zig"},
	}
	for lang, want := range tests {
		if got := RequiredBinaries(lang); !slices.Equal(got, want) {
			t.Errorf("RequiredBinaries(%q) = %q, want %q", lang, got, want)
		}
	}
	if got := RequiredBinaries(language.Language("unknown")); got != nil {
		t.Fatalf("unknown language binaries = %q, want nil", got)
	}
}

func TestProbeRecordsAvailableToolsInLanguageOrder(t *testing.T) {
	var lookedUp, inspected []string
	results := Probe([]language.Language{language.TypeScript, language.Go}, func(binary string) (string, error) {
		lookedUp = append(lookedUp, binary)
		return "/tools/" + binary, nil
	}, func(binary, path string) (string, error) {
		inspected = append(inspected, binary+":"+path)
		return binary + " 1.0", nil
	})
	if !slices.Equal(lookedUp, []string{"tsc", "node", "go"}) ||
		!slices.Equal(inspected, []string{"tsc:/tools/tsc", "node:/tools/node", "go:/tools/go"}) {
		t.Fatalf("looked up = %q, inspected = %q", lookedUp, inspected)
	}
	if len(results) != 2 || results[0].Language != language.TypeScript || results[1].Language != language.Go ||
		len(results[0].Tools) != 2 || len(results[1].Tools) != 1 ||
		results[0].Tools[0] != (Tool{Binary: "tsc", Path: "/tools/tsc", Version: "tsc 1.0"}) ||
		results[0].Tools[1] != (Tool{Binary: "node", Path: "/tools/node", Version: "node 1.0"}) ||
		results[1].Tools[0] != (Tool{Binary: "go", Path: "/tools/go", Version: "go 1.0"}) {
		t.Fatalf("results = %+v", results)
	}
}

func TestProbeReportsMissingAndUnusableTools(t *testing.T) {
	results := Probe([]language.Language{language.TypeScript}, func(binary string) (string, error) {
		if binary == "node" {
			return "", errors.New("not found")
		}
		return "/tools/" + binary, nil
	}, func(binary, path string) (string, error) {
		return "", errors.New("unusable")
	})
	if len(results) != 1 || len(results[0].Tools) != 2 ||
		results[0].Tools[0].Binary != "tsc" || results[0].Tools[0].Path != "/tools/tsc" || results[0].Tools[0].Error == nil ||
		results[0].Tools[1].Binary != "node" || results[0].Tools[1].Path != "" || results[0].Tools[1].Error == nil {
		t.Fatalf("results = %+v", results)
	}
}
