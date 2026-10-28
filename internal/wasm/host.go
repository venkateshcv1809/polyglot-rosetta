//go:build js && wasm

package wasm

import (
	"fmt"
	"syscall/js"
)

var (
	refreshCache    js.Value
	ensureAvailable js.Value
	execute         js.Value
	clearCache      js.Value
)

// Bind returns the JavaScript binding that registers required host callbacks.
func Bind() Binding {
	return Binding{Name: "bind", Handler: bind}
}

func bind(_ js.Value, args []js.Value) any {
	if len(args) != 1 || args[0].Type() != js.TypeObject || args[0].IsNull() {
		return map[string]any{"error": "ProsettaCore.bind requires {refreshCache, ensureAvailable, execute}"}
	}
	host := args[0]
	nextRefreshCache := host.Get("refreshCache")
	nextEnsureAvailable := host.Get("ensureAvailable")
	nextExecute := host.Get("execute")
	nextClearCache := host.Get("clearCache")
	if nextRefreshCache.Type() != js.TypeFunction || nextEnsureAvailable.Type() != js.TypeFunction || nextExecute.Type() != js.TypeFunction {
		return map[string]any{"error": "ProsettaCore.bind requires refreshCache, ensureAvailable, and execute functions"}
	}
	if nextClearCache.Type() != js.TypeUndefined && nextClearCache.Type() != js.TypeFunction {
		return map[string]any{"error": "ProsettaCore.bind clearCache must be a function"}
	}
	refreshCache, ensureAvailable, execute = nextRefreshCache, nextEnsureAvailable, nextExecute
	clearCache = nextClearCache
	return map[string]any{"ok": true}
}

// RefreshCache returns the callback that inspects toolchains and replaces the readiness snapshot.
func RefreshCache() (js.Value, error) {
	return required("refreshCache", refreshCache)
}

// EnsureAvailable returns the callback that makes a language ready using its cached snapshot.
func EnsureAvailable() (js.Value, error) {
	return required("ensureAvailable", ensureAvailable)
}

// Execute returns the bound worker/execution callback.
func Execute() (js.Value, error) {
	return required("execute", execute)
}

// ClearCache returns the optional callback that deletes the readiness snapshot.
func ClearCache() (js.Value, error) {
	if clearCache.Type() != js.TypeFunction {
		return js.Undefined(), fmt.Errorf("ProsettaCore.bind requires a clearCache callback to clear the probe snapshot")
	}
	return clearCache, nil
}

func required(name string, value js.Value) (js.Value, error) {
	if value.Type() != js.TypeFunction {
		return js.Undefined(), fmt.Errorf("ProsettaCore.bind({refreshCache, ensureAvailable, execute}) is required before %s", name)
	}
	return value, nil
}
