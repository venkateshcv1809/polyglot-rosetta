package probecmd

import (
	"errors"
	"strings"
	"testing"

	probeapi "polyglot-rosetta/internal/api/probe"
)

func TestProbeFiltersLanguageAndKeepsHostError(t *testing.T) {
	host := staticHost{rows: []probeapi.Toolchain{
		{Language: "go", Available: true, Tools: []probeapi.Tool{{Binary: "go", Available: true}}},
		{Language: "python", Available: false, Tools: []probeapi.Tool{{Binary: "python3", Available: false}}},
	}}
	result := Probe(" Go ", host)
	if result.Status != probeapi.StatusOK || len(result.Toolchains) != 1 || result.Toolchains[0].Language != "go" {
		t.Fatalf("filtered result = %+v", result)
	}

	failed := Probe("go", staticHost{rows: host.rows, err: errors.New("save toolchain cache: disk full")})
	if failed.Status != probeapi.StatusError || !strings.Contains(failed.Error, "save toolchain cache") || len(failed.Toolchains) != 1 {
		t.Fatalf("host error result = %+v", failed)
	}
}

func TestProbeRejectsUnsupportedAndAbsentLanguages(t *testing.T) {
	result := Probe("ruby", staticHost{})
	if result.Status != probeapi.StatusError || !strings.Contains(result.Error, `unsupported language "ruby"`) {
		t.Fatalf("unsupported = %+v", result)
	}

	absent := Probe("python", staticHost{rows: []probeapi.Toolchain{
		{Language: "go", Available: true, Tools: []probeapi.Tool{{Binary: "go", Available: true}}},
	}})
	if absent.Status != probeapi.StatusError || !strings.Contains(absent.Error, `language "python" is absent from the runtime manifest`) {
		t.Fatalf("absent = %+v", absent)
	}
}

type clearFunc func() error

func (clear clearFunc) ClearCache() error { return clear() }

func TestClearSnapshotResult(t *testing.T) {
	for _, hostErr := range []error{nil, errors.New("cannot delete snapshot")} {
		calls := 0
		result := ClearCache(clearFunc(func() error { calls++; return hostErr }))
		if calls != 1 || len(result.Toolchains) != 0 {
			t.Fatalf("clear calls = %d, result = %+v", calls, result)
		}
		if hostErr == nil {
			if result.Status != probeapi.StatusCleared || result.Error != "" {
				t.Fatalf("clear result = %+v", result)
			}
		} else if result.Status != probeapi.StatusError || result.Error != hostErr.Error() {
			t.Fatalf("failed clear = %+v", result)
		}
	}
}
