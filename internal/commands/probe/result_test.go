package probecmd

import (
	"errors"
	"testing"

	probeapi "polyglot-rosetta/internal/api/probe"
	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/toolchain"
)

func TestResultFromToolchain(t *testing.T) {
	result := resultFromToolchain([]toolchain.Result{
		{Language: language.Go, Tools: []toolchain.Tool{{Binary: "go", Path: "/bin/go", Version: "go 1.0"}}},
		{Language: language.TypeScript, Tools: []toolchain.Tool{
			{Binary: "tsc", Path: "/bin/tsc", Error: errors.New("exit status 1")},
			{Binary: "node", Error: errors.New("not found")},
		}},
	})
	if result.Status != probeapi.StatusUnavailable || result.Error != errToolchainsUnavailable.Error() || len(result.Toolchains) != 2 {
		t.Fatalf("result = %+v", result)
	}
	goTool := result.Toolchains[0].Tools[0]
	if !result.Toolchains[0].Available || !goTool.Available || goTool.Version != "go 1.0" || goTool.Path != "/bin/go" {
		t.Fatalf("Go toolchain = %+v", result.Toolchains[0])
	}
	ts := result.Toolchains[1]
	if ts.Available || ts.Tools[0].Available || ts.Tools[0].Path != "/bin/tsc" || ts.Tools[0].Error != "exit status 1" || ts.Tools[1].Error != "not found" {
		t.Fatalf("TypeScript toolchain = %+v", ts)
	}
}
