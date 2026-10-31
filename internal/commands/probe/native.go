package probecmd

import (
	probeapi "polyglot-rosetta/internal/api/probe"
	"polyglot-rosetta/internal/toolchain"
)

// NativeHost inspects PATH binaries and replaces the toolchain cache.
type NativeHost struct {
	CachePath string
	LookPath  func(string) (string, error)
	Inspect   func(string, string) (string, error)
}

func (host NativeHost) ClearCache() error {
	return toolchain.ClearCache(host.CachePath)
}

func (host NativeHost) RefreshCache() ([]probeapi.Toolchain, error) {
	results, err := toolchain.RefreshCache(host.CachePath, host.LookPath, host.Inspect)
	return nativeToolchains(results), err
}

func nativeToolchains(results []toolchain.Result) []probeapi.Toolchain {
	rows := make([]probeapi.Toolchain, 0, len(results))
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
			}
			entry.Tools = append(entry.Tools, item)
		}
		rows = append(rows, entry)
	}
	return rows
}
