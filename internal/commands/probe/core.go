package probecmd

import (
	"fmt"

	probeapi "polyglot-rosetta/internal/api/probe"
	"polyglot-rosetta/internal/domain/language"
)

// Host refreshes the readiness snapshot and returns the inspected toolchains.
type Host interface {
	RefreshCache() ([]probeapi.Toolchain, error)
}

// Probe reports language support using the supplied host. language is an
// optional filter; empty means every toolchain the host returns.
func Probe(languageName string, host Host) probeapi.Result {
	selected := ""
	if languageName != "" {
		parsed, err := language.Parse(languageName)
		if err != nil {
			return errorResult(fmt.Errorf("unsupported language %q", languageName))
		}
		selected = string(parsed)
	}
	rows, hostErr := host.RefreshCache()
	result := probeapi.NewResult(probeapi.StatusOK)
	for _, entry := range rows {
		if selected != "" && entry.Language != selected {
			continue
		}
		result.Toolchains = append(result.Toolchains, entry)
		if !entry.Available {
			result.Status = probeapi.StatusUnavailable
			result.Error = errToolchainsUnavailable.Error()
		}
	}
	if selected != "" && len(result.Toolchains) == 0 && hostErr == nil {
		result.Status = probeapi.StatusError
		result.Error = fmt.Sprintf("language %q is absent from the runtime manifest", selected)
		return result
	}
	if hostErr != nil {
		result.Status = probeapi.StatusError
		result.Error = hostErr.Error()
	}
	return result
}

// ClearHost removes only the saved readiness snapshot, never language tools.
type ClearHost interface {
	ClearCache() error
}

// ClearCache removes the host's probe snapshot without inspecting toolchains.
func ClearCache(host ClearHost) probeapi.Result {
	if err := host.ClearCache(); err != nil {
		return errorResult(err)
	}
	return probeapi.NewResult(probeapi.StatusCleared)
}
