//go:build js && wasm

// Package wasm defines JavaScript bindings exposed by the WASM entry point.
package wasm

import "syscall/js"

// Binding associates a JavaScript property name with a Go handler.
// Handler receives the JavaScript receiver and arguments and returns a JS-compatible value.
type Binding struct {
	Name    string
	Handler func(this js.Value, args []js.Value) any
}
