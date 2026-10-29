//go:build js && wasm

package wasm

import (
	"fmt"
	"strings"
	"syscall/js"

	"polyglot-rosetta/internal/api"
)

// A successful bind replaces the entire map; failed validation changes nothing.
var boundHost = map[string]js.Value{}

type invalidOption struct {
	Name        string `json:"name"`
	Expected    string `json:"expected"`
	Received    string `json:"received"`
	Description string `json:"description"`
}

type bindResult struct {
	SchemaVersion   int             `json:"schema_version"`
	Status          string          `json:"status"`
	OK              bool            `json:"ok"`
	Error           string          `json:"error,omitempty"`
	MissingRequired []string        `json:"missing_required,omitempty"`
	InvalidOptions  []invalidOption `json:"invalid_options,omitempty"`
	Options         []Option        `json:"options,omitempty"`
	Help            string          `json:"help,omitempty"`
}

// Bind registers the host callbacks using the requirements in hostOptions.
func Bind() Binding {
	return Binding{Name: "bind", Handler: bind, Help: bindHelp}
}

func bind(_ js.Value, args []js.Value) any {
	result := bindResult{SchemaVersion: api.SchemaVersion, Status: "error"}
	next := make(map[string]js.Value, len(hostOptions))
	if len(args) != 1 || args[0].Type() != js.TypeObject || args[0].IsNull() || js.Global().Get("Array").Call("isArray", args[0]).Bool() {
		received := "no arguments"
		if len(args) == 1 {
			received = valueType(args[0])
		} else if len(args) > 1 {
			received = fmt.Sprintf("%d arguments", len(args))
		}
		result.InvalidOptions = []invalidOption{{Name: "host", Expected: "one non-null object", Received: received, Description: "Supply the callback object documented by ProsettaCore.help(\"bind\")."}}
	} else {
		for _, option := range hostOptions {
			value := args[0].Get(option.Name)
			if value.Type() == js.TypeUndefined {
				if option.Required {
					result.MissingRequired = append(result.MissingRequired, option.Name)
				}
				continue
			}
			if value.Type() != js.TypeFunction {
				result.InvalidOptions = append(result.InvalidOptions, invalidOption{Name: option.Name, Expected: option.Type, Received: valueType(value), Description: option.Description})
				continue
			}
			next[option.Name] = value
		}
	}
	if len(result.MissingRequired) > 0 || len(result.InvalidOptions) > 0 {
		details := []string{}
		if len(result.MissingRequired) > 0 {
			details = append(details, "missing required callbacks: "+strings.Join(result.MissingRequired, ", "))
		}
		for _, invalid := range result.InvalidOptions {
			details = append(details, fmt.Sprintf("%s: expected %s, received %s", invalid.Name, invalid.Expected, invalid.Received))
		}
		result.Error = "ProsettaCore.bind: " + strings.Join(details, "; ") + ". Call ProsettaCore.help(\"bind\") for callback descriptions and signatures."
		result.Options, result.Help = hostOptions, "ProsettaCore.help(\"bind\")"
		return Value(result)
	}
	boundHost = next
	result.Status, result.OK = "ok", true
	return Value(result)
}

func valueType(value js.Value) string {
	if value.IsNull() {
		return "null"
	}
	if value.Type() == js.TypeObject && js.Global().Get("Array").Call("isArray", value).Bool() {
		return "array"
	}
	return value.Type().String()
}

// RefreshCache returns the callback that inspects toolchains and replaces the readiness snapshot.
func RefreshCache() (js.Value, error) { return hostCallback("refreshCache") }

// EnsureAvailable returns the callback that makes a language ready using its cached snapshot.
func EnsureAvailable() (js.Value, error) { return hostCallback("ensureAvailable") }

// Execute returns the bound worker/execution callback.
func Execute() (js.Value, error) { return hostCallback("execute") }

// ClearCache returns the optional callback that deletes the readiness snapshot.
func ClearCache() (js.Value, error) { return hostCallback("clearCache") }

func hostCallback(name string) (js.Value, error) {
	value := boundHost[name]
	if value.Type() != js.TypeFunction {
		return js.Undefined(), fmt.Errorf("host callback %q is not bound; call ProsettaCore.bind with this function before using it; see ProsettaCore.help(\"bind\")", name)
	}
	return value, nil
}
