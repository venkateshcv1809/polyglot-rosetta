//go:build js && wasm

package main

import (
	"encoding/hex"
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
		if missing.Get("missing_required").Length() != 5 || missing.Get("options").Length() != help.Get("host_options").Length() {
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

	t.Run("staged execution and core grading", func(t *testing.T) {
		f.activate(t)
		builds, runs, disposals := 0, 0, 0
		var program js.Value
		f.bind(t, map[string]any{
			"buildProgram": f.callback(t, func(args []js.Value) any {
				builds++
				program = args[0]
				if program.Get("language").String() != "typescript" || program.Get("entry").String() != "main.ts" || program.Get("files").Get("solution.ts").String() != "export class Solution {}" {
					t.Error("normalized source program missing")
				}
				if strings.Contains(jsonString(program), `"expected"`) {
					t.Error("expected answers were sent to host")
				}
				return buildResult(args, "")
			}),
			"runCase": f.callback(t, func(args []js.Value) any {
				runs++
				r := args[0]
				if r.Get("args").Index(0).String() != fmt.Sprint(r.Get("index").Int()) || r.Get("timeout_ms").Int() != 10000 {
					t.Error("invalid case launch")
				}
				return caseResult(args, greeting(r.Get("index").Int()))
			}),
			"disposeProgram": f.callback(t, func(args []js.Value) any {
				disposals++
				if args[0].String() != program.Get("session_id").String() {
					t.Error("wrong session disposed")
				}
				return nil
			}),
		})
		result := await(t, core.Call("run", request()))
		checkStatus(t, result, "passed")
		checkString(t, result.Get("id"), f.id)
		checkString(t, result.Get("concept"), "fundamentals/hello-world")
		if builds != 1 || runs != 1 || disposals != 1 || result.Get("build_duration_ms").Int() != 7 {
			t.Fatal("wrong lifecycle or timing")
		}
		all := request()
		all["mode"] = "all"
		result = await(t, core.Call("run", all))
		checkStatus(t, result, "passed")
		if builds != 2 || runs != 3 || disposals != 2 || result.Get("total_passed").Int() != 2 || !result.Get("test_results").Index(1).Get("hidden").Bool() {
			t.Fatal("wrong public/hidden selection")
		}
	})

	t.Run("readiness and invalid requests stop before build", func(t *testing.T) {
		f.activate(t)
		builds := 0
		f.bind(t, map[string]any{
			"ensureAvailable": f.callback(t, func([]js.Value) any {
				return wasm.Value(map[string]any{"schema_version": 1, "languages": map[string]any{}})
			}),
			"buildProgram": f.callback(t, func([]js.Value) any { builds++; return js.Null() }),
		})
		checkFailure(t, await(t, core.Call("run", request())), "not ready")
		f.bind(t, map[string]any{"buildProgram": f.callback(t, func([]js.Value) any { builds++; return js.Null() })})
		for _, r := range []any{js.Null(), []any{}, map[string]any{"id": f.id, "language": "go", "source": ""}, map[string]any{"id": f.id, "language": "ruby", "source": "x"}} {
			checkFailure(t, await(t, core.Call("run", r)), "")
		}
		if builds != 0 {
			t.Fatal("built invalid or unready request")
		}
	})

	t.Run("case failures do not stop later cases", func(t *testing.T) {
		all := request()
		all["mode"] = "all"
		for _, kind := range []string{"wrong_answer", "crashed", "timeout", "missing record"} {
			t.Run(kind, func(t *testing.T) {
				runs, disposals := 0, 0
				f.bind(t, map[string]any{
					"runCase": f.callback(t, func(args []js.Value) any {
						runs++
						index := args[0].Get("index").Int()
						r := caseResult(args, greeting(index))
						if index == 0 {
							switch kind {
							case "wrong_answer":
								r = caseResult(args, `"wrong"`)
							case "crashed":
								r.Set("exit_code", 2)
							case "timeout":
								r.Set("timed_out", true)
								r.Set("exit_code", -1)
							case "missing record":
								r.Set("stdout", "debug only")
							}
						}
						return r
					}),
					"disposeProgram": f.callback(t, func([]js.Value) any { disposals++; return nil }),
				})
				result := await(t, core.Call("run", all))
				checkStatus(t, result, "failed")
				expected := kind
				if kind == "missing record" {
					expected = "crashed"
				}
				checkString(t, result.Get("test_results").Index(0).Get("status"), expected)
				checkString(t, result.Get("test_results").Index(1).Get("status"), "passed")
				if runs != 2 || disposals != 1 || result.Get("total_passed").Int() != 1 {
					t.Fatal("lost subsequent case or disposal")
				}
			})
		}
	})

	t.Run("failed and invalid builds always dispose", func(t *testing.T) {
		mutations := []struct {
			name, want string
			change     func(js.Value)
		}{
			{"compiler failure", "compile failed", func(r js.Value) { r.Set("error", "compile failed") }},
			{"schema", "invalid build result", func(r js.Value) { r.Set("schema_version", 2) }},
			{"null schema", "schema_version", func(r js.Value) { r.Set("schema_version", js.Null()) }},
			{"nonfinite duration", "build_duration_ms", func(r js.Value) { r.Set("build_duration_ms", js.Global().Get("NaN")) }},
			{"session", "invalid build result", func(r js.Value) { r.Set("session_id", "wrong") }},
			{"negative duration", "invalid build result", func(r js.Value) { r.Set("build_duration_ms", -1) }},
			{"missing duration", "build_duration_ms", func(r js.Value) { js.Global().Get("Reflect").Call("deleteProperty", r, "build_duration_ms") }},
		}
		for _, mutation := range mutations {
			t.Run(mutation.name, func(t *testing.T) {
				runs, disposals := 0, 0
				f.bind(t, map[string]any{
					"buildProgram":   f.callback(t, func(args []js.Value) any { r := buildResult(args, ""); mutation.change(r); return r }),
					"runCase":        f.callback(t, func([]js.Value) any { runs++; return nil }),
					"disposeProgram": f.callback(t, func([]js.Value) any { disposals++; return nil }),
				})
				result := await(t, core.Call("run", request()))
				checkFailure(t, result, mutation.want)
				if runs != 0 || disposals != 1 || result.Get("test_results").Length() != 0 {
					t.Fatal("invalid build lifecycle")
				}
				if mutation.name == "compiler failure" && result.Get("build_duration_ms").Int() != 7 {
					t.Fatal("failed build timing lost")
				}
			})
		}
	})

	t.Run("bridge errors and invalid case results abort and dispose", func(t *testing.T) {
		cases := []struct {
			name, want string
			callback   any
		}{
			{"synchronous throw", "browser execution failed", js.Global().Get("JSON").Get("parse")},
			{"rejection", "backend rejected", f.callback(t, func([]js.Value) any { return rejected("backend rejected") })},
			{"undefined", "JSON object", f.callback(t, func([]js.Value) any { return js.Undefined() })},
			{"wrong session", "invalid case result", f.callback(t, func(args []js.Value) any { r := caseResult(args, greeting(0)); r.Set("session_id", "wrong"); return r })},
			{"wrong index", "invalid case result", f.callback(t, func(args []js.Value) any { r := caseResult(args, greeting(0)); r.Set("index", 9); return r })},
			{"wrong schema", "invalid case result", f.callback(t, func(args []js.Value) any { r := caseResult(args, greeting(0)); r.Set("schema_version", 2); return r })},
			{"invalid exit", "exit_code", f.callback(t, func(args []js.Value) any { r := caseResult(args, greeting(0)); r.Set("exit_code", -2); return r })},
			{"null exit", "exit_code", f.callback(t, func(args []js.Value) any { r := caseResult(args, greeting(0)); r.Set("exit_code", js.Null()); return r })},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				disposals := 0
				f.bind(t, map[string]any{"runCase": c.callback, "disposeProgram": f.callback(t, func([]js.Value) any { disposals++; return nil })})
				checkFailure(t, await(t, core.Call("run", request())), c.want)
				if disposals != 1 {
					t.Fatal("session leaked")
				}
			})
		}
		disposals := 0
		f.bind(t, map[string]any{"buildProgram": f.callback(t, func([]js.Value) any { return rejected("build rejected") }), "disposeProgram": f.callback(t, func([]js.Value) any { disposals++; return rejected("cleanup failed") })})
		result := await(t, core.Call("run", request()))
		checkFailure(t, result, "build rejected")
		if !strings.Contains(result.Get("error").String(), "cleanup failed") || disposals != 1 {
			t.Fatal("cleanup error was lost")
		}
	})

	t.Run("concurrent runs and pinned host callbacks", func(t *testing.T) {
		sessions := map[string]bool{}
		f.bind(t, map[string]any{"buildProgram": f.callback(t, func(args []js.Value) any {
			id := args[0].Get("session_id").String()
			if sessions[id] {
				t.Error("session reused")
			}
			sessions[id] = true
			return hostPromise(func() js.Value { time.Sleep(time.Millisecond); return buildResult(args, "") })
		})})
		var pending []js.Value
		for i := 0; i < 8; i++ {
			pending = append(pending, core.Call("run", request()))
		}
		for _, p := range pending {
			checkStatus(t, await(t, p), "passed")
		}
		if len(sessions) != 8 {
			t.Fatal("sessions collided")
		}
		var calls []string
		started, release := make(chan struct{}), make(chan struct{})
		bind := func(label string, block bool) {
			f.bind(t, map[string]any{
				"ensureAvailable": f.callback(t, func([]js.Value) any {
					calls = append(calls, label+".ready")
					if block {
						return hostPromise(func() js.Value { close(started); <-release; return readySnapshot() })
					}
					return readySnapshot()
				}),
				"buildProgram": f.callback(t, func(args []js.Value) any { calls = append(calls, label+".build"); return buildResult(args, "") }),
				"runCase": f.callback(t, func(args []js.Value) any {
					calls = append(calls, label+".case")
					return caseResult(args, greeting(args[0].Get("index").Int()))
				}),
				"disposeProgram": f.callback(t, func([]js.Value) any { calls = append(calls, label+".dispose"); return nil }),
			})
		}
		bind("A", true)
		a := core.Call("run", request())
		<-started
		bind("B", false)
		b := await(t, core.Call("run", request()))
		close(release)
		ar := await(t, a)
		checkStatus(t, b, "passed")
		checkStatus(t, ar, "passed")
		if !reflect.DeepEqual(calls, []string{"A.ready", "B.ready", "B.build", "B.case", "B.dispose", "A.build", "A.case", "A.dispose"}) {
			t.Fatalf("host calls=%v", calls)
		}
	})

	t.Run("catalog refresh preserves pending run", func(t *testing.T) {
		f.activate(t)
		release := make(chan struct{})
		f.bind(t, map[string]any{"ensureAvailable": f.callback(t, func([]js.Value) any { return hostPromise(func() js.Value { <-release; return readySnapshot() }) })})
		pending := core.Call("run", request())
		refreshed := core.Call("catalog", f.index)
		close(release)
		checkStatus(t, refreshed, "ok")
		checkStatus(t, await(t, pending), "passed")
	})

	t.Run("structured answers are compared in Go", func(t *testing.T) {
		f.loadBundleWithCases(t, `[
            {"id":1,"input":{},"expected":{"values":[true]},"hidden":false},
            {"id":2,"input":{},"expected":{"label":"😀","matrix":[[1,2]],"none":null},"hidden":true}
        ]`)
		f.activate(t)
		f.bind(t, map[string]any{"runCase": f.callback(t, func(args []js.Value) any {
			if args[0].Get("index").Int() == 0 {
				return caseResult(args, `{"values":[1]}`)
			}
			return caseResult(args, "{\n\"none\":null,\"matrix\":[[1,2]],\"label\":\"\\ud83d\\ude00\"\n}")
		})})
		all := request()
		all["mode"] = "all"
		result := await(t, core.Call("run", all))
		checkStatus(t, result, "failed")
		checkString(t, result.Get("test_results").Index(0).Get("status"), "wrong_answer")
		checkString(t, result.Get("test_results").Index(1).Get("status"), "passed")
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
		"buildProgram":    f.callback(t, func(args []js.Value) any { return buildResult(args, "") }),
		"runCase":         f.callback(t, func(args []js.Value) any { return caseResult(args, greeting(args[0].Get("index").Int())) }),
		"disposeProgram":  f.callback(t, func([]js.Value) any { return nil }),
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
	f.loadBundleWithCases(t, "")
}

func (f *wasmFixture) loadBundleWithCases(t *testing.T, cases string) {
	t.Helper()
	output := t.TempDir()
	problems, templates := testfixture.Catalog(t)
	if cases != "" {
		if err := os.WriteFile(filepath.Join(problems, "fundamentals/hello-world/testcases.json"), []byte(cases), 0o600); err != nil {
			t.Fatal(err)
		}
	}
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

func buildResult(args []js.Value, message string) js.Value {
	program := args[0]
	result := map[string]any{"schema_version": 1, "session_id": program.Get("session_id").String()}
	if program.Get("build_required").Bool() {
		result["build_duration_ms"] = 7
	}
	if message != "" {
		result["error"] = message
	}
	return wasm.Value(result)
}

func caseResult(args []js.Value, jsonOutput string) js.Value {
	request := args[0]
	session := request.Get("session_id").String()
	index := request.Get("index").Int()
	return wasm.Value(map[string]any{
		"schema_version": 1, "session_id": session, "index": index, "exit_code": 0, "timed_out": false, "stderr": "",
		"stdout": fmt.Sprintf("debug log\n%s:%d:1000:json:%s\n", session, index, hex.EncodeToString([]byte(jsonOutput))),
	})
}

func greeting(index int) string {
	if index == 1 {
		return `"Hello Test"`
	}
	return `"Hello World"`
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
