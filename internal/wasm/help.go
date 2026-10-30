//go:build js && wasm

package wasm

import (
	"fmt"
	"strings"
	"syscall/js"

	"polyglot-rosetta/internal/api"
)

// Option describes an argument or host field. Required defaults to false.
type Option struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Description string   `json:"description"`
	Default     any      `json:"default,omitempty"`
	Values      []string `json:"values,omitempty"`
	Properties  []Option `json:"properties,omitempty"`
	Signature   string   `json:"signature,omitempty"`
}

// MethodHelp documents one method exposed by this particular WASM entry point.
type MethodHelp struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Async       bool     `json:"async"`
	Signatures  []string `json:"signatures"`
	Parameters  []Option `json:"parameters"`
	Callbacks   []string `json:"callbacks"`
	Returns     string   `json:"returns"`
	Notes       []string `json:"notes,omitempty"`
}

type helpResult struct {
	SchemaVersion int          `json:"schema_version"`
	Status        string       `json:"status"`
	Methods       []MethodHelp `json:"methods"`
	HostOptions   []Option     `json:"host_options"`
	Error         string       `json:"error,omitempty"`
}

// Required is set only for mandatory callbacks; omitted means optional.
// Validation and help both read this configuration.
var hostOptions = []Option{
	{Name: "refreshCache", Type: "function", Required: true,
		Description: "Inspect browser toolchains and replace the readiness snapshot; used by probe.",
		Signature:   "() => {toolchains: ProbeToolchain[]} | Promise<{toolchains: ProbeToolchain[]}>"},
	{Name: "ensureAvailable", Type: "function", Required: true,
		Description: "Make a language ready and return its readiness snapshot; used by run.",
		Signature:   "(language: string) => CacheSnapshot | Promise<CacheSnapshot>"},
	{Name: "execute", Type: "function", Required: true,
		Description: "Compile and evaluate the normalized request against the selected test cases; used by run.",
		Signature:   "(request: RunRequest, testCases: TestCase[]) => RunResult | Promise<RunResult>"},
	{Name: "clearCache", Type: "function",
		Description: "Delete only the readiness snapshot; required when probe is called with clear: true. Preserve compiler assets.",
		Signature:   "() => void | Promise<void>"},
}

var bindHelp = MethodHelp{
	Description: "Validate and register browser callbacks. Failed binding preserves the previous host.",
	Signatures:  []string{"bind(host)"},
	Parameters:  []Option{{Name: "host", Type: "object", Required: true, Description: "Callback object; see host_options for required fields, signatures, and optional fields."}},
	Returns:     "{schema_version, status, ok, error?, missing_required?, invalid_options?, options?, help?}",
	Notes:       []string{"Required callbacks must all be supplied, even when using only probe. Optional callbacks must be functions if supplied.", "Unknown fields are ignored. Callbacks may return immediate values or promises."},
}

var helpMethod = MethodHelp{
	Description: "Describe available methods and required or optional host callbacks.",
	Signatures:  []string{"help()", "help(method)"},
	Parameters:  []Option{{Name: "method", Type: "string", Description: "An optional method name; omit it to list every registered method."}},
	Returns:     "{schema_version, status, methods, host_options, error?}",
}

// Help builds discovery from the same bindings that main registers. No host
// callbacks, catalog, or network access are required to ask for help.
func Help(bindings []Binding) Binding {
	self := helpMethod
	methods := make([]MethodHelp, 0, len(bindings)+1)
	for _, binding := range bindings {
		info := binding.Help
		info.Name = binding.Name
		methods = append(methods, normalizeHelp(info))
	}
	self.Name = "help"
	methods = append(methods, normalizeHelp(self))
	return Binding{Name: "help", Help: self, Handler: func(_ js.Value, args []js.Value) any {
		result := helpResult{SchemaVersion: api.SchemaVersion, Status: "ok", Methods: methods, HostOptions: hostOptions}
		if len(args) > 1 || len(args) == 1 && args[0].Type() != js.TypeString {
			result.Status, result.Error = "error", "help accepts no arguments or one method name; call ProsettaCore.help() to list methods"
			return Value(result)
		}
		if len(args) == 1 {
			name := args[0].String()
			for _, method := range methods {
				if method.Name == name {
					result.Methods = []MethodHelp{method}
					return Value(result)
				}
			}
			names := make([]string, 0, len(methods))
			for _, method := range methods {
				names = append(names, method.Name)
			}
			result.Status, result.Error = "error", fmt.Sprintf("unknown method %q; available methods: %s", name, strings.Join(names, ", "))
		}
		return Value(result)
	}}
}

func normalizeHelp(info MethodHelp) MethodHelp {
	if info.Signatures == nil {
		info.Signatures = []string{}
	}
	if info.Parameters == nil {
		info.Parameters = []Option{}
	}
	if info.Callbacks == nil {
		info.Callbacks = []string{}
	}
	return info
}
