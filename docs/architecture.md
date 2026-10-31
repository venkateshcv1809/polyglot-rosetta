# Architecture and browser integration

[Project README](../README.md#documentation)

Polyglot Rosetta uses a shared Go command core for the native CLI and browser integrations. Each host supplies catalog access, toolchain readiness, and solution execution.

## System design

```mermaid
flowchart TB
    Author[Exercise authors] --> Generate[Maintainer CLI]
    Generate --> Content[Problems and shared templates]
    Content --> Bundle[Catalog bundle]
    Bundle --> Published[Published catalog]

    Terminal[Terminal or editor] --> CLI[Native CLI]
    App[Browser application] --> SDK[Host integration]
    CLI --> Core[Shared Go command core]
    SDK --> WASM[Go WASM bindings]
    WASM --> Core

    Content --> NativeCatalog[Local catalog or activated cache]
    Published --> NativeCatalog
    NativeCatalog --> Core
    Published --> SDK
    SDK --> BrowserCatalog[Verified catalog in memory]
    BrowserCatalog --> Core

    Core --> Native[Native execution host]
    Native --> Harness[Temporary harness and system toolchains]
    Core --> Browser[Browser host callbacks]
    Browser --> Workers[Workers and browser toolchain assets]
    Harness --> Results[Versioned JSON results]
    Workers --> Results
    Results --> Core
```

The native and browser execution paths are alternatives selected by the adapter. The engine and catalog do not include language compilers: native runs use tools on `PATH`, while a browser host supplies its own runtime assets. The browser application, workers, and runtime delivery are integration responsibilities; a browser SDK is not shipped in this repository.

## Evaluation flow

```mermaid
sequenceDiagram
    participant Client as CLI or browser client
    participant Core as Shared command core
    participant Catalog as Catalog
    participant Host as Execution host
    Client->>Core: Concept ID, language, source, mode
    Core->>Core: Validate and normalize request
    Core->>Catalog: Resolve concept and select cases
    Catalog-->>Core: Public or all test cases
    Core->>Host: Ensure toolchain is available
    Host-->>Core: Readiness
    Core->>Host: Execute source against selected cases
    Host-->>Core: Case results and timing
    Core->>Core: Validate result and attach catalog metadata
    Core-->>Client: Structured run result
```

The native host generates a harness in a temporary workspace, builds when needed, and runs each case in a separate process. It bounds execution time and captured output, then cleans up the workspace. Submissions execute with the user's access; this is trusted local practice, not a sandbox.

The browser host must implement readiness and execution callbacks and provide worker management and deadlines. The Go core handles request validation, case selection, host invocation, and result validation in both environments. It delegates compilation and case execution to the host.

## Integration boundaries

Editors send the current buffer through `prosetta run --stdin --format=json` and preserve the process exit status. `prosetta catalog` returns the local content root for reading exercises and index metadata.

Browser hosts load `prosetta.wasm` with the matching `wasm_exec.js`, verify downloaded catalog assets before activation, and bind their readiness and execution callbacks. Runtime-specific cross-origin isolation requirements and language runtime delivery belong to the browser host deployment. See the [browser implementation guide](browser-integration.md) for the contract and activation sequence.

The JavaScript entry point is `ProsettaCore`. Call `help()` to discover methods or `help("bind")` for host callback requirements. Hosts bind `refreshCache`, `ensureAvailable`, and `execute`; `clearCache` is needed when clearing readiness. `probe` and `run` return Promises, while `version`, `catalog`, and `help` return synchronously.

JSON requests and results are defined in [`internal/api`](../internal/api/). Clients should validate `schema_version` (currently `1`) and use full concept IDs. See the [development guide](development.md#local-setup-and-builds) for build commands.

## Source map

- [`cmd/prosetta`](../cmd/prosetta/): public native entry point.
- [`cmd/generate`](../cmd/generate/): maintainer entry point for scaffolding, discovery, and packaging.
- [`cmd/wasm`](../cmd/wasm/): browser exports registered on `ProsettaCore`.
- [`internal/commands`](../internal/commands/): command services and native/WASM adapters.
- [`internal/api`](../internal/api/): transport-neutral request and result types.
- [`internal/catalog`](../internal/catalog/): metadata, IDs, content enrichment, bundles, and synchronization.
- [`internal/execution`](../internal/execution/): native harness generation and process execution.
- [`internal/toolchain`](../internal/toolchain/): native tool discovery and readiness cache.
- [`internal/wasm`](../internal/wasm/): JavaScript conversion, promises, discovery, and host binding.

The Go module uses the local name `polyglot-rosetta`, and most packages are internal. Integrate through the CLI or WASM surface; this repository does not advertise a supported public Go library API.
