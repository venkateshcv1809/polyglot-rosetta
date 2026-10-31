//go:build js && wasm

package wasm

import (
	"reflect"
	"strings"
	"syscall/js"
	"testing"
)

func testHost(t *testing.T) map[string]any {
	t.Helper()
	original := boundHost
	boundHost = map[string]js.Value{}
	t.Cleanup(func() { boundHost = original })
	callback := js.FuncOf(func(js.Value, []js.Value) any { return nil })
	t.Cleanup(callback.Release)
	return map[string]any{"refreshCache": callback, "ensureAvailable": callback, "execute": callback}
}

func bindForTest(t *testing.T, args ...js.Value) bindResult {
	t.Helper()
	var result bindResult
	if err := Decode(bind(js.Undefined(), args).(js.Value), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestBindReportsAllMissingAndInvalidOptions(t *testing.T) {
	testHost(t)
	result := bindForTest(t, js.ValueOf(map[string]any{}))
	if result.OK || result.Status != "error" || !reflect.DeepEqual(result.MissingRequired, []string{"refreshCache", "ensureAvailable", "execute"}) {
		t.Fatalf("result=%+v", result)
	}
	if len(result.Options) != len(hostOptions) || !strings.Contains(result.Error, "help(\"bind\")") {
		t.Fatalf("missing discovery details: %+v", result)
	}
	result = bindForTest(t, js.ValueOf(map[string]any{"refreshCache": "wrong", "clearCache": false}))
	if len(result.InvalidOptions) != 2 || !reflect.DeepEqual(result.MissingRequired, []string{"ensureAvailable", "execute"}) {
		t.Fatalf("result=%+v", result)
	}
	if result.InvalidOptions[0].Received != "string" || result.InvalidOptions[1].Received != "boolean" {
		t.Fatalf("invalid types: %+v", result.InvalidOptions)
	}
	for _, name := range []string{"refreshCache", "ensureAvailable", "execute", "clearCache"} {
		if !strings.Contains(result.Error, name) {
			t.Fatalf("error omits %s: %s", name, result.Error)
		}
	}
}

func TestBindRejectsInvalidArgumentShape(t *testing.T) {
	testHost(t)
	for _, args := range [][]js.Value{nil, {js.Null()}, {js.ValueOf([]any{})}, {js.ValueOf("host")}, {js.ValueOf(map[string]any{}), js.ValueOf(map[string]any{})}} {
		result := bindForTest(t, args...)
		if result.OK || len(result.InvalidOptions) != 1 || result.InvalidOptions[0].Name != "host" || len(result.Options) != len(hostOptions) {
			t.Fatalf("result=%+v", result)
		}
	}
}

func TestBindPreservesPreviousHostOnFailureAndClearsOmittedOptional(t *testing.T) {
	host := testHost(t)
	host["clearCache"] = host["refreshCache"]
	host["label"] = "ignored metadata"
	if r := bindForTest(t, js.ValueOf(host)); !r.OK || r.Status != "ok" {
		t.Fatalf("result=%+v", r)
	}
	originalRefresh := boundHost["refreshCache"]
	originalClear := boundHost["clearCache"]
	host["clearCache"] = 42
	if r := bindForTest(t, js.ValueOf(host)); r.OK {
		t.Fatal("invalid optional callback accepted")
	}
	if !boundHost["refreshCache"].Equal(originalRefresh) || !boundHost["clearCache"].Equal(originalClear) {
		t.Fatal("failed bind changed host")
	}
	delete(host, "clearCache")
	if r := bindForTest(t, js.ValueOf(host)); !r.OK {
		t.Fatalf("omitted optional rejected: %+v", r)
	}
	if _, err := ClearCache(); err == nil || !strings.Contains(err.Error(), "help(\"bind\")") {
		t.Fatalf("missing optional callback error=%v", err)
	}
	if _, err := RefreshCache(); err != nil {
		t.Fatal(err)
	}
}

func TestBindAndHelpUseTheSameRequiredConfiguration(t *testing.T) {
	host := testHost(t)
	originalOptions := hostOptions
	t.Cleanup(func() { hostOptions = originalOptions })
	hostOptions = append(append([]Option{}, hostOptions...), Option{Name: "extra", Type: "function", Required: true, Description: "Test configured requirement."})
	result := bindForTest(t, js.ValueOf(host))
	if !reflect.DeepEqual(result.MissingRequired, []string{"extra"}) {
		t.Fatalf("new config not validated: %+v", result)
	}
	info := helpForTest(t, Help([]Binding{Bind()}))
	if info.HostOptions[len(info.HostOptions)-1].Name != "extra" || !info.HostOptions[len(info.HostOptions)-1].Required {
		t.Fatal("help did not use host config")
	}
	host["extra"] = host["execute"]
	if r := bindForTest(t, js.ValueOf(host)); !r.OK {
		t.Fatalf("configured function rejected: %+v", r)
	}
}
