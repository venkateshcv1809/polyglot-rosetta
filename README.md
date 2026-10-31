# Polyglot Rosetta

**Learn a concept once. Implement it across five languages.**

Polyglot Rosetta brings programming exercises, reference implementations, and local evaluation into one workflow for **Go, Python, Rust, TypeScript, and Zig**. Each exercise pairs a precise problem statement with public and hidden test cases, an explanation, and single-file reference solutions.

The `prosetta` CLI evaluates solutions with your installed toolchains. A shared Go WebAssembly core gives browser applications access to the evaluation contract through host-provided runtimes.

## What you can build with it

- **Cross-language practice:** solve the same exercise in different languages and compare the implementation choices.
- **Editor workflows:** evaluate the current buffer through standard input and consume versioned JSON results.
- **Custom learning experiences:** author exercises from shared templates and distribute a catalog independently of the engine.
- **Browser integrations:** embed the WASM core and supply toolchain readiness, compilation, and execution callbacks.

## Project status

The project is in the `v0.x` stage. The current catalog contains a Hello World exercise in all five languages, demonstrating the complete authoring and evaluation workflow. The engine supports local execution, catalog packaging and synchronization, and browser host bindings. This repository does not include a hosted practice application or a packaged browser SDK.

Use it for trusted practice code and integration development. Native execution has time and output limits but runs with your account's permissions. See the [security policy](SECURITY.md) before designing a service around it.

## Quick start

Install the Go version specified in [go.mod](go.mod) and [just](https://just.systems/), then run these commands from the repository root:

```sh
just build-cli
just list
./bin/prosetta probe --lang=go
./bin/prosetta run --id=c-2b23ca0786e0eb237f64cf2b8759bcc3 --lang=go
```

The last command runs the Hello World reference solution against its public tests. To evaluate your own solution, add `--stdin < solution.go`; add `--all` to include hidden cases.

See the [getting started guide](docs/getting-started.md) for toolchain setup and the complete practice workflow.

## Documentation

- [Getting started](docs/getting-started.md)
- [CLI and configuration](docs/cli.md)
- [Writing exercises](docs/authoring.md)
- [Architecture and browser integration](docs/architecture.md)
- [Browser host implementation](docs/browser-integration.md)
- [Development, testing, and releases](docs/development.md)
- [Contributing](CONTRIBUTING.md)
- [Support and bug reports](SUPPORT.md)
- [Security policy](SECURITY.md)
- [Public release readiness](docs/release-readiness.md)

## License and availability

The code and documentation are provided under the [MIT License](LICENSE). Repository access and release downloads remain subject to the repository's current visibility and your access permissions. Paid support, service-level agreements, and a managed hosting service are not offered by this repository.
