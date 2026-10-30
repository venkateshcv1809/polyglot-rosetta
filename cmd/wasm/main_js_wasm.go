//go:build js && wasm

// Command wasm exposes application operations to JavaScript hosts.
package main

import (
	"syscall/js"

	probecmd "polyglot-rosetta/internal/commands/probe"
	versioncmd "polyglot-rosetta/internal/commands/version"
	"polyglot-rosetta/internal/wasm"
)

func main() {
	core := js.Global().Get("Object").New()
	bindings := []wasm.Binding{
		wasm.Bind(),
		probecmd.WASM(),
		versioncmd.WASM(),
	}
	bindings = append(bindings, wasm.Help(bindings))
	for _, binding := range bindings {
		callback := js.FuncOf(binding.Handler)
		defer callback.Release()
		core.Set(binding.Name, callback)
	}

	js.Global().Set("ProsettaCore", core)
	select {}
}
