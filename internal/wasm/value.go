//go:build js && wasm

package wasm

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"polyglot-rosetta/internal/api"
)

// Decode converts a JavaScript object to a typed Go request.
func Decode(value js.Value, target any) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("decode JavaScript value: %v", recovered)
		}
	}()
	if value.Type() != js.TypeObject || value.IsNull() || js.Global().Get("Array").Call("isArray", value).Bool() {
		return fmt.Errorf("expected a JSON object")
	}
	encoded := js.Global().Get("JSON").Call("stringify", value)
	if encoded.Type() != js.TypeString {
		return fmt.Errorf("value must be JSON serializable")
	}
	return json.Unmarshal([]byte(encoded.String()), target)
}

// Value converts a Go result to a JavaScript object using its JSON field names.
func Value(result any) js.Value {
	encoded, err := json.Marshal(result)
	if err != nil {
		return js.ValueOf(map[string]any{
			"schema_version": api.SchemaVersion, "status": "error", "error": err.Error(),
		})
	}
	return js.Global().Get("JSON").Call("parse", string(encoded))
}

// Bytes copies an archive or index from JavaScript into Go-owned memory.
func Bytes(value js.Value) ([]byte, error) {
	if !value.InstanceOf(js.Global().Get("Uint8Array")) || value.Length() == 0 || value.Length() > 128<<20 {
		return nil, fmt.Errorf("expected a non-empty Uint8Array of at most 128 MiB")
	}
	data := make([]byte, value.Length())
	if copied := js.CopyBytesToGo(data, value); copied != len(data) {
		return nil, fmt.Errorf("incomplete byte copy")
	}
	return data, nil
}
