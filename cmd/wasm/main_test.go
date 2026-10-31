//go:build js && wasm

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"syscall/js"
	"testing"
	"time"

	"polyglot-rosetta/internal/catalog/bundle"
	"polyglot-rosetta/internal/testfixture"
	"polyglot-rosetta/internal/version"
	"polyglot-rosetta/internal/wasm"
)

// These tests execute the entry point and its real JavaScript exports in a Go
// test WASM binary, rather than loading the standalone release artifact.
func TestPublicWASMBindings(t *testing.T) {
	go main()
	core := js.Global().Get("ProsettaCore")
	for core.IsUndefined() {
		time.Sleep(time.Millisecond)
		core = js.Global().Get("ProsettaCore")
	}
	f := &wasmFixture{core: core}

	t.Run("exports and discovery before binding", func(t *testing.T) {
		names := js.Global().Get("Object").Call("keys", core)
		got := make([]string, names.Length())
		for i := range got {
			got[i] = names.Index(i).String()
		}
		sort.Strings(got)
		if !reflect.DeepEqual(got, []string{"bind", "catalog", "help", "probe", "run", "version"}) {
			t.Fatalf("exports = %v", got)
		}
		for _, name := range []string{"Prosetta", "ProsettaWASM"} {
			if !js.Global().Get(name).IsUndefined() {
				t.Errorf("unexpected legacy global %s", name)
			}
		}
		checkStatus(t, core.Call("version"), "ok")
		checkString(t, core.Call("version").Get("version"), version.Value)
		help := core.Call("help")
		checkStatus(t, help, "ok")
		if help.Get("methods").Length() != len(got) {
			t.Fatal("help does not describe all exports")
		}
		runHelp := core.Call("help", "run").Get("methods").Index(0)
		properties := runHelp.Get("parameters").Index(0).Get("properties")
		var required []string
		for i := 0; i < properties.Length(); i++ {
			property := properties.Index(i)
			if property.Get("required").Truthy() {
				required = append(required, property.Get("name").String())
			}
			if property.Get("name").String() == "mode" {
				checkString(t, property.Get("default"), "public")
			}
		}
		if !reflect.DeepEqual(required, []string{"id", "language", "source"}) {
			t.Fatalf("required run properties = %v", required)
		}
		if !core.Call("help", "probe").Get("methods").Index(0).Get("async").Bool() {
			t.Fatal("probe is not documented as asynchronous")
		}
		checkFailure(t, core.Call("help", "unknown"), "available methods")
		checkFailure(t, core.Call("help", 7), "method name")
		missing := core.Call("bind", map[string]any{})
		checkFailure(t, missing, "refreshCache")
		if missing.Get("missing_required").Length() != 3 || missing.Get("options").Length() != help.Get("host_options").Length() {
			t.Fatal("bind discovery is incomplete")
		}
		checkFailure(t, core.Call("version", "extra"), "accepts no arguments")
		checkFailure(t, core.Call("catalog"), "not loaded")
		checkFailure(t, core.Call("catalog", js.Null()), "Uint8Array")
		checkFailure(t, core.Call("catalog", js.Null(), js.Null(), js.Null()), "catalog accepts")
		checkFailure(t, core.Call("catalog", js.Null(), js.Null(), map[string]any{}), "Uint8Array")
		checkFailure(t, await(t, core.Call("probe")), "bind")
		checkFailure(t, await(t, core.Call("run", map[string]any{"id": "c-" + strings.Repeat("0", 32), "language": "go", "source": "x"})), "bind")
	})

	t.Run("probe and cache clearing", func(t *testing.T) {
		toolchains := []any{
			map[string]any{"language": "go", "available": false, "tools": []any{map[string]any{"binary": "go.wasm", "available": false, "error": "requires an isolated page"}}},
			map[string]any{"language": "rust", "available": false, "tools": []any{map[string]any{"binary": "rust", "available": false}}},
			map[string]any{"language": "typescript", "available": true, "tools": []any{map[string]any{"binary": "typescript.js", "available": true}}},
		}
		refresh := f.callback(t, func([]js.Value) any {
			return js.Global().Get("Promise").Call("resolve", wasm.Value(map[string]any{"toolchains": toolchains}))
		})
		f.bind(t, map[string]any{"refreshCache": refresh})
		probe := await(t, core.Call("probe"))
		checkStatus(t, probe, "unavailable")
		if probe.Get("toolchains").Length() != 3 || !strings.Contains(probe.Get("toolchains").Index(0).Get("tools").Index(0).Get("error").String(), "isolated") {
			t.Fatal("probe lost the host's unavailable toolchains")
		}
		checkStatus(t, await(t, core.Call("probe", " TypeScript ")), "ok")
		toolchains[0].(map[string]any)["available"] = true
		toolchains[0].(map[string]any)["tools"].([]any)[0].(map[string]any)["available"] = true
		checkStatus(t, await(t, core.Call("probe", "go")), "ok")
		for _, query := range []struct {
			value any
			error string
		}{{"python", "absent"}, {"ruby", "unsupported"}, {7, "must be a string"}} {
			checkFailure(t, await(t, core.Call("probe", query.value)), query.error)
		}
		clear := map[string]any{"clear": true}
		checkFailure(t, await(t, core.Call("probe", clear)), "clearCache")
		clears, refreshes := 0, 0
		f.bind(t, map[string]any{
			"refreshCache": f.callback(t, func([]js.Value) any { refreshes++; return js.Null() }),
			"clearCache":   f.callback(t, func([]js.Value) any { clears++; return js.Global().Get("Promise").Call("resolve") }),
		})
		for i := 0; i < 2; i++ {
			result := await(t, core.Call("probe", clear))
			checkStatus(t, result, "cleared")
			if result.Get("toolchains").Length() != 0 {
				t.Fatal("clear returned toolchains")
			}
		}
		checkFailure(t, await(t, core.Call("probe", map[string]any{"clear": true, "language": "go"})), "cannot be combined")
		if clears != 2 || refreshes != 0 {
			t.Fatalf("clear calls=%d refresh calls=%d", clears, refreshes)
		}
		f.bind(t, map[string]any{"clearCache": f.callback(t, func([]js.Value) any { return rejected("snapshot storage blocked") })})
		checkFailure(t, await(t, core.Call("probe", clear)), "snapshot storage blocked")
		f.bind(t, nil)
		checkFailure(t, await(t, core.Call("probe", clear)), "clearCache")
	})

	f.loadBundle(t)
	request := func() map[string]any {
		return map[string]any{"id": f.id[:10], "language": " TypeScript ", "source": "export class Solution {}"}
	}

	t.Run("catalog activation and rollback", func(t *testing.T) {
		f.bind(t, nil)
		checkFailure(t, await(t, core.Call("run", request())), "not loaded")
		checkFailure(t, await(t, core.Call("run")), "requires a request")
		checkStatus(t, core.Call("catalog", f.index), "ok")
		checkString(t, core.Call("catalog", f.id[:10]).Get("concept_id"), f.id)
		checkFailure(t, core.Call("catalog", "c-"+strings.Repeat("0", 32)), "not found")
		checkFailure(t, await(t, core.Call("run", request())), "no selected test cases")
		loaded := core.Call("catalog", f.index, f.templates, f.archives)
		checkStatus(t, loaded, "ok")
		starter := loaded.Get("catalog").Get("concepts_by_id").Get(f.id).Get("starters").Get("typescript").String()
		if !strings.Contains(starter, "Solution") {
			t.Fatal("template starter missing")
		}
		badArchives := js.Global().Get("Object").New()
		badArchives.Set("fundamentals.tar.gz", byteArray([]byte{1}))
		checkFailure(t, core.Call("catalog", f.index, f.templates, badArchives), "decode")
		if jsonString(core.Call("catalog")) != jsonString(loaded) {
			t.Fatal("failed activation changed catalog")
		}
	})

	t.Run("execution readiness and case selection", func(t *testing.T) {
		f.activate(t)
		calls := 0
		f.bind(t, map[string]any{
			"ensureAvailable": f.callback(t, func([]js.Value) any {
				return wasm.Value(map[string]any{"schema_version": 1, "languages": map[string]any{}})
			}),
			"execute": f.callback(t, func([]js.Value) any { calls++; return js.Null() }),
		})
		checkFailure(t, await(t, core.Call("run", request())), "not ready")
		if calls != 0 {
			t.Fatal("execute called before readiness")
		}
		var normalized, selected js.Value
		f.bind(t, map[string]any{"execute": f.callback(t, func(args []js.Value) any {
			normalized, selected = args[0], args[1]
			return backendResult(args, "passed")
		})})
		result := await(t, core.Call("run", request()))
		checkStatus(t, result, "passed")
		checkString(t, normalized.Get("id"), f.id)
		checkString(t, normalized.Get("language"), "typescript")
		checkString(t, normalized.Get("mode"), "public")
		checkString(t, result.Get("concept"), "fundamentals/hello-world")
		if selected.Length() != 1 || selected.Index(0).Get("hidden").Bool() || result.Get("total_passed").Int() != 1 {
			t.Fatal("wrong public cases")
		}
		all := request()
		all["mode"] = "all"
		result = await(t, core.Call("run", all))
		checkStatus(t, result, "passed")
		if selected.Length() != 2 || result.Get("test_results").Index(0).Get("hidden").Bool() || !result.Get("test_results").Index(1).Get("hidden").Bool() {
			t.Fatal("hidden cases not preserved")
		}
		for _, status := range []string{"failed", "timeout", "error"} {
			f.bind(t, map[string]any{"execute": f.callback(t, func(args []js.Value) any { return backendResult(args, status) })})
			result := await(t, core.Call("run", request()))
			checkStatus(t, result, status)
			if result.Get("total_failed").Int() != 1 {
				t.Fatal("wrong failed count")
			}
		}
	})

	t.Run("invalid requests and backend failures", func(t *testing.T) {
		f.activate(t)
		calls := 0
		f.bind(t, map[string]any{"execute": f.callback(t, func([]js.Value) any { calls++; return js.Null() })})
		invalid := []any{js.Null(), []any{}}
		for key, value := range map[string]any{"source": "", "mode": "other", "language": "ruby", "id": "c-" + strings.Repeat("0", 32)} {
			r := request()
			r[key] = value
			invalid = append(invalid, r)
		}
		for _, r := range invalid {
			checkFailure(t, await(t, core.Call("run", r)), "")
		}
		if calls != 0 {
			t.Fatal("backend called with invalid request")
		}
		// JSON.parse throws synchronously when called with the request object.
		f.bind(t, map[string]any{"execute": js.Global().Get("JSON").Get("parse")})
		checkFailure(t, await(t, core.Call("run", request())), "browser execution failed")
		f.bind(t, map[string]any{"execute": f.callback(t, func([]js.Value) any { return rejected("async backend failure") })})
		checkFailure(t, await(t, core.Call("run", request())), "async backend failure")
		f.bind(t, map[string]any{"execute": f.callback(t, func([]js.Value) any { return js.Undefined() })})
		checkFailure(t, await(t, core.Call("run", request())), "JSON object")
		cyclic := js.Global().Get("Object").New()
		cyclic.Set("languages", cyclic)
		f.bind(t, map[string]any{"execute": f.callback(t, func([]js.Value) any { return cyclic })})
		checkFailure(t, await(t, core.Call("run", request())), "decode JavaScript value")
		changes := []struct {
			name   string
			change func(js.Value)
		}{
			{"schema", func(r js.Value) { r.Set("schema_version", 2) }},
			{"identity", func(r js.Value) { r.Set("id", "different") }},
			{"counts", func(r js.Value) { r.Set("total_passed", 0) }},
			{"duration", func(r js.Value) { r.Set("duration_ms", -1) }},
			{"missing cases", func(r js.Value) { r.Set("test_results", []any{}) }},
			{"case passed", func(r js.Value) { r.Get("test_results").Index(0).Set("passed", false) }},
			{"case status", func(r js.Value) { r.Get("test_results").Index(0).Set("status", "unknown") }},
			{"status", func(r js.Value) { r.Set("status", "unknown") }},
			{"timeout consistency", func(r js.Value) { r.Set("status", "timeout") }},
		}
		for _, change := range changes {
			t.Run(change.name, func(t *testing.T) {
				f.bind(t, map[string]any{"execute": f.callback(t, func(args []js.Value) any { r := backendResult(args, "passed"); change.change(r); return r })})
				checkFailure(t, await(t, core.Call("run", request())), "invalid execution result")
			})
		}
	})

	t.Run("concurrent runs and pinned host callbacks", func(t *testing.T) {
		f.activate(t)
		f.bind(t, map[string]any{"execute": f.callback(t, func(args []js.Value) any {
			return hostPromise(func() js.Value { time.Sleep(time.Millisecond); return backendResult(args, "passed") })
		})})
		var pending []js.Value
		for i := 0; i < 8; i++ {
			pending = append(pending, core.Call("run", request()))
		}
		for _, promise := range pending {
			checkStatus(t, await(t, promise), "passed")
		}
		var calls []string
		started, release := make(chan struct{}), make(chan struct{})
		f.bind(t, map[string]any{
			"ensureAvailable": f.callback(t, func([]js.Value) any {
				calls = append(calls, "A.ensureAvailable")
				return hostPromise(func() js.Value { close(started); <-release; return readySnapshot() })
			}),
			"execute": f.callback(t, func(args []js.Value) any { calls = append(calls, "A.execute"); return backendResult(args, "passed") }),
		})
		runA := core.Call("run", request())
		<-started
		f.bind(t, map[string]any{
			"ensureAvailable": f.callback(t, func([]js.Value) any { calls = append(calls, "B.ensureAvailable"); return readySnapshot() }),
			"execute":         f.callback(t, func(args []js.Value) any { calls = append(calls, "B.execute"); return backendResult(args, "passed") }),
		})
		// Always unblock A, even if the assertion for B fails.
		resultB := await(t, core.Call("run", request()))
		close(release)
		resultA := await(t, runA)
		checkStatus(t, resultB, "passed")
		checkStatus(t, resultA, "passed")
		if !reflect.DeepEqual(calls, []string{"A.ensureAvailable", "B.ensureAvailable", "B.execute", "A.execute"}) {
			t.Fatalf("host calls = %v", calls)
		}
	})

	t.Run("catalog refresh preserves pending run", func(t *testing.T) {
		f.activate(t)
		release := make(chan struct{})
		f.bind(t, map[string]any{"ensureAvailable": f.callback(t, func([]js.Value) any {
			return hostPromise(func() js.Value { <-release; return readySnapshot() })
		})})
		pending := core.Call("run", request())
		refreshed := core.Call("catalog", f.index)
		close(release)
		checkStatus(t, refreshed, "ok")
		checkStatus(t, await(t, pending), "passed")
		checkString(t, core.Call("version").Get("version"), version.Value)
	})
}

type wasmFixture struct {
	core, index, templates, archives js.Value
	id                               string
}

func (f *wasmFixture) callback(t *testing.T, fn func([]js.Value) any) js.Func {
	t.Helper()
	callback := js.FuncOf(func(_ js.Value, args []js.Value) (result any) {
		defer func() {
			if recovered := recover(); recovered != nil {
				result = rejected(fmt.Sprintf("test host callback panicked: %v", recovered))
			}
		}()
		return fn(args)
	})
	t.Cleanup(callback.Release)
	return callback
}

func (f *wasmFixture) bind(t *testing.T, overrides map[string]any) {
	t.Helper()
	host := map[string]any{
		"refreshCache":    f.callback(t, func([]js.Value) any { return wasm.Value(map[string]any{"toolchains": []any{}}) }),
		"ensureAvailable": f.callback(t, func([]js.Value) any { return readySnapshot() }),
		"execute":         f.callback(t, func(args []js.Value) any { return backendResult(args, "passed") }),
	}
	for name, callback := range overrides {
		host[name] = callback
	}
	result := f.core.Call("bind", host)
	if !result.Get("ok").Bool() {
		t.Fatalf("bind failed: %s", jsonString(result))
	}
}

func (f *wasmFixture) loadBundle(t *testing.T) {
	t.Helper()
	output := t.TempDir()
	problems, templates := testfixture.Catalog(t)
	if err := bundle.Build(problems, templates, output); err != nil {
		t.Fatal(err)
	}
	read := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(output, name))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	index := read("index.json")
	metadata, err := bundle.ParseIndex(index)
	if err != nil {
		t.Fatal(err)
	}
	for id, concept := range metadata.ConceptsByID {
		if concept.Path == "fundamentals/hello-world" {
			f.id = id
		}
	}
	if f.id == "" {
		t.Fatal("test catalog must contain hello-world")
	}
	f.index, f.templates = byteArray(index), byteArray(read("templates.tar.gz"))
	manifest, err := bundle.ParseManifest(read("manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	f.archives = js.Global().Get("Object").New()
	for _, archive := range manifest.Categories {
		f.archives.Set(archive.Path, byteArray(read(archive.Path)))
	}
}

func (f *wasmFixture) activate(t *testing.T) {
	t.Helper()
	checkStatus(t, f.core.Call("catalog", f.index, f.templates, f.archives), "ok")
}

func readySnapshot() js.Value {
	return wasm.Value(map[string]any{"schema_version": 1, "languages": map[string]any{
		"typescript": map[string]any{"required": []string{"typescript.js"}, "available": true},
	}})
}

func backendResult(args []js.Value, status string) js.Value {
	request, tests := args[0], args[1]
	passed, failed, caseStatus := 0, tests.Length(), "wrong_answer"
	if status == "passed" {
		passed, failed, caseStatus = tests.Length(), 0, "passed"
	}
	if status == "timeout" {
		caseStatus = "timeout"
	}
	if status == "error" {
		caseStatus = "crashed"
	}
	rows := make([]any, tests.Length())
	for i := range rows {
		rows[i] = map[string]any{"id": tests.Index(i).Get("id").String(), "status": caseStatus, "passed": status == "passed", "hidden": false, "solution_time_ns": 1000}
	}
	return wasm.Value(map[string]any{"schema_version": 1, "id": request.Get("id").String(), "language": request.Get("language").String(), "mode": request.Get("mode").String(), "status": status, "total_passed": passed, "total_failed": failed, "duration_ms": 2, "test_results": rows})
}

func await(t *testing.T, promise js.Value) js.Value {
	t.Helper()
	if !promise.InstanceOf(js.Global().Get("Promise")) {
		t.Fatal("async export did not return a Promise")
	}
	identity := js.Global().Get("Object")
	result, err := wasm.Invoke(identity, promise)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// Resolve JavaScript values directly: wasm.Promise marshals Go result structs.
func hostPromise(work func() js.Value) js.Value {
	executor := js.FuncOf(func(_ js.Value, args []js.Value) any {
		resolve, reject := args[0], args[1]
		go func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					reject.Invoke(js.Global().Get("Error").New(fmt.Sprint(recovered)))
				}
			}()
			resolve.Invoke(work())
		}()
		return nil
	})
	defer executor.Release()
	return js.Global().Get("Promise").New(executor)
}

func rejected(message string) js.Value {
	return js.Global().Get("Promise").Call("reject", js.Global().Get("Error").New(message))
}
func jsonString(value js.Value) string {
	return js.Global().Get("JSON").Call("stringify", value).String()
}
func byteArray(data []byte) js.Value {
	value := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(value, data)
	return value
}
func checkStatus(t *testing.T, result js.Value, status string) {
	t.Helper()
	if result.Get("schema_version").Int() != 1 || result.Get("status").String() != status {
		t.Fatalf("want status %s: %s", status, jsonString(result))
	}
}
func checkString(t *testing.T, value js.Value, want string) {
	t.Helper()
	if value.Type() != js.TypeString || value.String() != want {
		t.Fatalf("got %v, want %q", value, want)
	}
}
func checkFailure(t *testing.T, result js.Value, message string) {
	t.Helper()
	checkStatus(t, result, "error")
	if text := result.Get("error").String(); text == "" || !strings.Contains(text, message) {
		t.Fatalf("want error containing %q: %s", message, jsonString(result))
	}
}
