package probecmd

import (
	probeapi "polyglot-rosetta/internal/api/probe"
	"polyglot-rosetta/internal/toolchain"
)

func resultFromToolchain(results []toolchain.Result) probeapi.Result {
	result := probeapi.Result{Status: probeapi.StatusOK, Toolchains: make([]probeapi.Toolchain, 0, len(results))}
	for _, inspected := range results {
		entry := probeapi.Toolchain{
			Language:  string(inspected.Language),
			Available: true,
			Tools:     make([]probeapi.Tool, 0, len(inspected.Tools)),
		}
		for _, tool := range inspected.Tools {
			item := probeapi.Tool{
				Binary: tool.Binary, Available: tool.Error == nil,
				Path: tool.Path, Version: tool.Version,
			}
			if tool.Error != nil {
				item.Error = tool.Error.Error()
				entry.Available = false
				result.Status = probeapi.StatusUnavailable
				result.Error = errToolchainsUnavailable.Error()
			}
			entry.Tools = append(entry.Tools, item)
		}
		result.Toolchains = append(result.Toolchains, entry)
	}
	return result
}
