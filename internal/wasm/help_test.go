//go:build js && wasm

package wasm

import (
	executionapi "polyglot-rosetta/internal/api/execution"
	"reflect"
	"strings"
	"syscall/js"
	"testing"
)

func helpForTest(t *testing.T, binding Binding, args ...js.Value) helpResult {
	t.Helper()
	var result helpResult
	if err := Decode(binding.Handler(js.Undefined(), args).(js.Value), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestExecutionContractDiscoveryMatchesWireTypes(t *testing.T) {
	for name, value := range map[string]any{
		"Program": executionapi.Program{}, "BuildResult": executionapi.BuildResult{},
		"CaseRequest": executionapi.CaseRequest{}, "CaseResult": executionapi.CaseResult{},
	} {
		typ := reflect.TypeOf(value)
		fields := hostContracts[name]
		if len(fields) != typ.NumField() {
			t.Fatalf("%s discovery fields do not match Go wire type", name)
		}
		for i, field := range fields {
			jsonTag := strings.Split(typ.Field(i).Tag.Get("json"), ",")
			optional := len(jsonTag) > 1 && jsonTag[1] == "omitempty"
			if field.Name != jsonTag[0] || field.Required == optional || field.Description == "" {
				t.Fatalf("%s field %d discovery is inconsistent: %+v", name, i, field)
			}
		}
	}
}

func TestHelpBeforeBindingAndSelectedMethod(t *testing.T) {
	testHost(t)
	binding := Help([]Binding{Bind(), {Name: "example", Help: MethodHelp{Description: "Example method", Signatures: []string{"example(request)"}, Parameters: []Option{{Name: "request", Type: "object", Required: true, Description: "Required request."}}, Returns: "Example result"}}})
	result := helpForTest(t, binding)
	names := []string{}
	for _, method := range result.Methods {
		names = append(names, method.Name)
		if method.Parameters == nil || method.Callbacks == nil {
			t.Fatal("help lists must be arrays")
		}
	}
	if result.SchemaVersion != 1 || result.HostAPIVersion != 1 || result.Status != "ok" || !reflect.DeepEqual(names, []string{"bind", "example", "help"}) {
		t.Fatalf("help=%+v", result)
	}
	if len(result.HostOptions) != 6 || !result.HostOptions[0].Required || result.HostOptions[5].Required {
		t.Fatalf("host options=%+v", result.HostOptions)
	}
	selected := helpForTest(t, binding, js.ValueOf("example"))
	if len(selected.Methods) != 1 || !selected.Methods[0].Parameters[0].Required || selected.Methods[0].Description != "Example method" {
		t.Fatalf("selected=%+v", selected)
	}
	if _, err := RefreshCache(); err == nil {
		t.Fatal("help unexpectedly bound a host")
	}
}

func TestHelpReportsInvalidQueries(t *testing.T) {
	binding := Help([]Binding{Bind()})
	for _, args := range [][]js.Value{{js.ValueOf(7)}, {js.ValueOf("bind"), js.ValueOf("extra")}, {js.ValueOf("unknown")}} {
		result := helpForTest(t, binding, args...)
		if result.Status != "error" || result.Error == "" || len(result.Methods) != 2 {
			t.Fatalf("result=%+v", result)
		}
	}
	unknown := helpForTest(t, binding, js.ValueOf("unknown"))
	if !strings.Contains(unknown.Error, "bind, help") {
		t.Fatalf("available names missing: %s", unknown.Error)
	}
}
