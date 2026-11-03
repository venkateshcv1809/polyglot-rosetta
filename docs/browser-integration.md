# Implementing a browser host

[Project overview](../README.md) · [Architecture](architecture.md) · [Security policy](../SECURITY.md)

The Go WASM core owns request validation, catalog resolution, case selection, harness generation, result parsing, answer comparison, and aggregation. A browser host supplies language runtimes and the three operations that depend on its environment: prepare/build a program, launch one case, and release the session. The native CLI uses the same evaluation pipeline with an operating-system host.

## Host contract

The host contract is still under development and has not been publicly released. Its initial `host_api_version` remains `1`, reported by `ProsettaCore.help()` and the result of `bind`. Command payloads and results use `schema_version: 1`.

Hosts implement `buildProgram`, `runCase`, and `disposeProgram`. Go generates the harnesses, compares actual and expected answers, and calculates pass/fail totals. The SDK should follow this staged contract rather than implement a separate whole-evaluation callback.

The public `run({id, language, source, mode?})` request and catalog operations remain the same. Runtime discovery includes callback signatures in `host_options` and wire fields in `host_contracts`. The authoritative Go wire types live in [`internal/api/execution`](../internal/api/execution/host.go).

## Initialize and bind

Build with `just build-wasm`, or obtain `prosetta.wasm` and `wasm_exec.js` from the same release. Load the matching Go loader, instantiate the module with the Go runtime's import object, and start it. Wait until `globalThis.ProsettaCore` is registered; the Go main function remains active, so do not wait for it to exit.

Serve WASM with `application/wasm` when using streaming instantiation. Configure CORS and any cross-origin isolation headers required by the chosen language runtimes. Runtime downloads, asset verification, worker management, and browser storage belong to the host.

`bind(host)` validates all callbacks before replacing the current host. Failed binding preserves the previous host. Required callbacks must be present even when initially using only `probe`; optional callbacks must be functions when supplied. Callbacks may return values or Promises.

Required callbacks:

- `refreshCache()`: inspect browser toolchains and replace readiness state; return `{toolchains: [...]}` using the [probe types](../internal/api/probe/result.go).
- `ensureAvailable(language)`: make the requested runtime ready; return a [cache snapshot](../internal/api/probe/cache.go) with `schema_version: 1` and `languages[language].available: true` when ready.
- `buildProgram(program)`: prepare the supplied files and compile once when required; return a `BuildResult`.
- `runCase(request)`: launch one case from the prepared session and return a raw `CaseResult`.
- `disposeProgram(sessionID)`: release all resources belonging to that session, including partial or failed builds. Return void.

`clearCache()` is optional until `probe({clear: true})` is used. It removes readiness state while preserving downloaded compiler assets.

Check `bind`'s `status`, `ok`, and `host_api_version` before continuing. Use `help("run")` and `help("bind")` for the running build's requirements. Bindings and the activated catalog are pinned at the start of each run: rebinding or refreshing the catalog cannot redirect a run already in progress.

## Build a session

After resolving the concept and checking readiness, Go generates the complete program and calls `buildProgram` once. `Program` contains:

- `schema_version: 1` and a unique, opaque `session_id` allocated by Go.
- Normalized `language`, the `entry` file, and a `files` map of filenames to source strings.
- `build_required`, false for Python and true for the compiled languages.
- `timeout_ms`, the host preparation/build deadline (60 seconds by default).

Register resources under `session_id` before starting asynchronous work. Compile or load the supplied program without rewriting its solution interface or result-reporting behavior. Store the compiled artifact or prepared runtime for subsequent case launches. Select browser-specific compiler targets and flags in the host; the core does not ask a browser to execute native shell commands.

A successful compiled build returns, for example:

```json
{
  "schema_version": 1,
  "session_id": "the-session-id-from-program",
  "build_duration_ms": 125
}
```

Return a nonnegative integer compiler duration for successful compiled programs, including zero for sub-millisecond builds. Include it on failed compiler attempts too, with an `error` diagnostic. Omit it for Python, whose runtime preparation is not a separately timed compilation step. The overall run duration still includes host preparation.

A build error stops evaluation before any case runs. The core still calls disposal with the allocated session ID. Returning an invalid build response or throwing/rejecting also triggers disposal.

## Run one case

Go calls `runCase` sequentially for the selected cases. Each `CaseRequest` contains `schema_version`, `session_id`, zero-based `index`, `args`, `timeout_ms`, and `max_output_bytes`. `args` excludes the executable name and currently contains the index as a string. This index is the position in the selected suite, not the authored testcase ID.

Launch the supplied entry in a fresh runtime environment for every request. Reuse the compiled artifact without reusing solution memory or module globals. A testcase can itself contain an operation sequence; all operations within that one call can share state.

Capture stdout and stderr separately. Return the request identity plus raw execution information:

```json
{
  "schema_version": 1,
  "session_id": "the-session-id-from-request",
  "index": 0,
  "stdout": "captured stdout, including the harness record",
  "stderr": "captured stderr",
  "exit_code": 0,
  "timed_out": false
}
```

Use `exit_code: -1` when no process exit code is available. `error` can carry a process/worker diagnostic; `stdout_truncated` and `stderr_truncated` indicate retained output tails. Enforce `max_output_bytes` per stream in bytes. Go also bounds the output it retains from host responses.

Return student crashes and deadlines as raw results (`exit_code`/`error` or `timed_out: true`). Go records that case's failure and continues with the next case. Throwing/rejecting a callback or returning a result with the wrong schema, session ID, or index signals a host infrastructure failure: the run aborts and the session is disposed.

Do not return grades, expected values, testcase totals, or synthetic pass records. Go parses the unchanged harness record and compares the returned value itself. Native and WASM evaluations therefore use the same comparison rules and solution interfaces.

## Runtime requirements

The generated program includes all selected inputs. Each launch selects its input through the supplied argument. Expected answers stay in Go; generated sources only encode the string-versus-JSON return convention needed by Rust and Zig. This is not a confidentiality boundary: hidden cases are distributed catalog content.

WASI-backed hosts must provide argument access, standard streams, and monotonic clocks. The Zig harness selects the WASI clock import when targeting WASI. Choose the supported toolchain version and supply any standard-library assets required by that compiler.

TypeScript hosts must compile both generated `main.ts` and `solution.ts` with CommonJS module resolution. When running outside Node, supply the generated entry's small Node-compatible surface: `process.argv[2]` for the case argument, `process.hrtime.bigint()` for monotonic nanoseconds, `TextEncoder`, and captured console output. Module loading must resolve `./solution`. Create a fresh module realm or worker for each case. TypeScript-specific compiler diagnostics remain a host responsibility.

## Deadlines and cleanup

The host must enforce case deadlines by terminating the case's worker/runtime and returning a timeout result. Preserve the session's compiled artifact so later cases can still run. The core allows one additional second for case-result delivery after the requested deadline. Readiness is bounded to 60 seconds, build/preparation to `Program.timeout_ms`, and disposal to five seconds.

Promise deadlines are guards against unresponsive asynchronous hosts; they cannot interrupt synchronous JavaScript or stop a worker themselves. Never run potentially blocking solution code on the core's JavaScript event loop. Late Promise settlement cannot change an already completed Go stage.

`disposeProgram` is called once after any attempted build, including failure. It must be safe for unknown or partially initialized sessions, cancel outstanding work, and prevent late build completion from recreating disposed resources. A disposal failure becomes a run error; if another error already occurred, both are retained.

## Catalog activation and requests

Fetch `manifest.json` from a trusted catalog endpoint. Validate its schema and paths, check `min_engine`, and verify asset sizes and SHA-256 values before activation. The browser core accepts bytes rather than a manifest, so the host owns those integrity and compatibility checks.

`catalog(indexBytes)` activates metadata only. `catalog(indexBytes, templateBytes, archives)` activates supplied categories, where each byte argument is a non-empty `Uint8Array` of at most 128 MiB and archive keys match manifest filenames. Activate the selected category before running it. Failed activation preserves the previous snapshot.

```js
const activated = ProsettaCore.catalog(indexBytes, templateBytes, {
  "fundamentals.tar.gz": fundamentalsBytes,
});
if (activated.status !== "ok") throw new Error(activated.error);

const result = await ProsettaCore.run({
  id: "c-2b23ca0786e0eb237f64cf2b8759bcc3",
  language: "go",
  source: solutionSource,
  mode: "public",
});
```

`mode` defaults to `public`; `all` includes hidden cases. `source` is text, not a path or URL. Validate `schema_version` and inspect the result's status; command failures normally resolve with `status: "error"`, while unexpected bridge failures can reject. Render diagnostics as text rather than HTML.

## Verification

The WASM suite exercises real exported bindings with fixture hosts: discovery, binding validation, host and catalog pinning, one build per run, sequential case launches, core grading, failure timings, cleanup, and malformed responses. It also checks Promise deadlines and late settlement. Native integration tests execute generated Go and Zig programs under Node's WASI host. These checks do not test the SDK's browser compilers, asset delivery, or worker implementation; verify those when implementing the SDK adapter.
