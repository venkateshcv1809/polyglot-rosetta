# Implementing a browser host

[Project overview](../README.md) · [Architecture](architecture.md) · [Security policy](../SECURITY.md)

The browser deliverable is a Go WASM core with a JavaScript callback interface. Your application supplies the UI, language runtimes, compilation, execution workers, deadlines, storage, and catalog downloads. A packaged JavaScript SDK and browser language runtimes are not included.

## Initialize the core

1. Build with `just build-wasm`, or obtain `prosetta.wasm` and `wasm_exec.js` from the same release. Keep the loader paired with its WASM artifact.
2. Load `wasm_exec.js`, instantiate the WASM module using the Go runtime's import object, and start the Go runtime. Its main function remains active; do not wait for it to exit before using the core.
3. Wait until `globalThis.ProsettaCore` is registered. Call `ProsettaCore.version()` and `ProsettaCore.help()` to check the running version and discover its contract.
4. Bind the host callbacks and activate catalog content before evaluating a solution.

Serve WASM with `application/wasm` when using streaming instantiation. Configure CORS for cross-origin assets and the security headers required by your chosen runtimes. These requirements depend on the host and runtime, not only the core.

## Bind host callbacks

`ProsettaCore.bind(host)` is synchronous. All three required callbacks must be present, even if the application initially uses only `probe`. Callbacks may return values or Promises:

- `refreshCache()`: inspect toolchains and replace the readiness snapshot; return `{toolchains: [...]}` using the [probe toolchain types](../internal/api/probe/result.go).
- `ensureAvailable(language)`: make the requested runtime ready; return a [cache snapshot](../internal/api/probe/cache.go) with `schema_version: 1` and `languages[language].available: true` when ready.
- `execute(request, testCases)`: compile and evaluate the supplied source against exactly the supplied cases; return the [run result](../internal/api/run/result.go).
- `clearCache()`: optional until `probe({clear: true})` is used; remove readiness state while preserving downloaded compiler assets.

Inspect `ProsettaCore.help("bind")` for the running build's callback signatures. Check the bind result's `status` and `ok` fields before continuing. Successful binding replaces the callback set; failed validation preserves the previous host. Runs capture their catalog and callback references before asynchronous work begins.

The host must enforce deadlines and terminate stalled workers. A callback Promise that never settles is not given an execution deadline by the core.

## Verify and activate catalog content

Fetch `manifest.json` from a trusted catalog deployment. Validate its schema and asset paths, check the engine against `min_engine`, and verify each downloaded asset's byte length and SHA-256 digest before activation. Browser catalog calls receive bytes rather than the manifest, so the core does not perform manifest digest or minimum-engine verification for you.

All byte arguments are non-empty `Uint8Array` values, at most 128 MiB each. The archive map uses the exact filenames from the manifest.

```js
// These values are downloaded and verified by your application.
const activated = ProsettaCore.catalog(indexBytes, templateArchiveBytes, {
  "fundamentals.tar.gz": fundamentalsArchiveBytes,
});
if (activated.status !== "ok") {
  throw new Error(activated.error);
}
```

`catalog(indexBytes)` activates metadata only, useful for discovery before downloading exercise content. `catalog(indexBytes, templateBytes, archives)` activates supplied category content. Activate the category containing the selected exercise before running it. A failed activation leaves the previous snapshot available.

Use `catalog()` to read the current snapshot, or `catalog(conceptID)` to resolve a full ID or unique prefix. Prefer full IDs in saved application state.

## Evaluate source

Once the host is bound and the exercise category is activated:

```js
const result = await ProsettaCore.run({
  id: "c-2b23ca0786e0eb237f64cf2b8759bcc3",
  language: "go",
  source: `package main
type Solution struct{}
func (Solution) Solve(input any) any {
    return "Hello " + input.(map[string]any)["name"].(string)
}`,
  mode: "public",
});
```

`source` is source text, not a URL or filename. `mode` defaults to `public`; `all` selects hidden cases too. The core normalizes the request, resolves the exercise, selects cases, calls readiness and execution, and validates the returned result.

The execution callback must return the normalized request identity (`id`, `language`, `mode`), `schema_version: 1`, nonnegative durations, consistent totals and status, and one result per selected case in the supplied order. Case IDs are strings. Per-case statuses are `passed`, `wrong_answer`, `crashed`, and `timeout`; the shared result contract accepts overall `passed`, `failed`, `timeout`, and `error` with consistency checks. See the [validator](../internal/commands/run/core.go) for exact requirements.

Command failures normally resolve to a result with `status: "error"`; unexpected bridge failures can reject the Promise. Handle both. Check `schema_version` before interpreting a response and render any error text as text rather than HTML.

## Probe and integration checks

`await ProsettaCore.probe()` refreshes readiness. A language filter limits the reported results; `await ProsettaCore.probe({clear: true})` uses the optional clear callback. `probe` and `run` are asynchronous; `bind`, `catalog`, `version`, and `help` are synchronous.

Validate your host against missing toolchains, wrong answers, compilation errors, crashes, worker deadlines, invalid callback results, failed catalog activation, and rebinding during an in-flight request. Run `just test-wasm` when changing the core; those fixture-based tests do not exercise your browser compiler or deployment.
