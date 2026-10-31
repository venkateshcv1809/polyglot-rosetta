# Getting started

[Project README](../README.md#documentation) · [CLI reference](cli.md)

## Requirements

To build the project, install the Go version declared in [`go.mod`](../go.mod) and `just`. To run solutions, install the toolchain for your chosen language:

- Go: `go`
- Python: `python3`
- Rust: `rustc`
- TypeScript: `tsc` and `node`
- Zig: `zig`

The binaries must be on `PATH`. You only need the solution toolchains you intend to use. The CLI and WASM core do not embed compilers.

If you use [mise](https://mise.jdx.dev/), the repository pins development tools in [`mise.toml`](../mise.toml) and [`mise.lock`](../mise.lock):

```sh
mise trust
mise install --locked
```

Run the commands below in an activated mise shell, or prefix them with `mise exec --`.

## Build and discover exercises

From the repository root:

```sh
just build-cli
./bin/prosetta version
just list
./bin/prosetta probe
```

`just list` prints concept paths, short IDs, full IDs, and titles from the live `problems/` tree. Use those IDs with `run`; paths and titles are not run IDs.

Run the included Hello World reference solution:

```sh
./bin/prosetta run --id=c-2b23ca0786e0eb237f64cf2b8759bcc3 --lang=go
```

This should pass the public greeting case. No catalog download is needed when running against this checkout's `problems/` tree.

## Practice with your own solution

Read the [Hello World problem](../problems/fundamentals/hello-world/PROBLEM.md), then copy the common Go starter to a working file:

```sh
mkdir -p practice
cp templates/concept/go/template.go practice/solution.go
```

Edit `practice/solution.go` to implement the required interface:

```go
package main

type Solution struct{}

func (Solution) Solve(input any) any {
    name := input.(map[string]any)["name"].(string)
    return "Hello " + name
}
```

Evaluate your source through standard input:

```sh
./bin/prosetta run --id=c-2b23ca0786e0eb237f64cf2b8759bcc3 --lang=go --stdin < practice/solution.go
./bin/prosetta run --id=c-2b23ca0786e0eb237f64cf2b8759bcc3 --lang=go --stdin --all < practice/solution.go
```

The first command runs public cases. The second includes hidden cases. Hidden cases are stored in the catalog; they are a practice distinction, not secret tests.

Keep practice files in their own directory or outside the repository so their Go package does not conflict with the root library package. Without `--stdin` or `--code`, `run` evaluates the catalog's reference solution. See [language interfaces](authoring.md#solution-interfaces) for the other languages.

## Use a release binary

Download the executable for your OS and architecture from the repository's [GitHub Releases](https://github.com/venkateshcv1809/polyglot-rosetta/releases), make it executable where required, and put it on `PATH`. Release assets include Linux, macOS, and Windows builds for `amd64` and `arm64`.

Release downloads require repository access while the repository is private. If no release is available to you, use the source build above. Download `SHA256SUMS` alongside the chosen asset and verify its checksum before installation. On Linux, use `sha256sum --ignore-missing --check SHA256SUMS` from the download directory; confirm the selected executable is explicitly reported as `OK`. On macOS, use `shasum -a 256 <downloaded-file>`; on Windows, use `Get-FileHash <downloaded-file> -Algorithm SHA256` in PowerShell, and compare with the corresponding entry in `SHA256SUMS`.

Binary names follow `prosetta-<version>-<os>-<arch>` (`.exe` on Windows); macOS uses `darwin`. Rename the installed binary to `prosetta` or `prosetta.exe` to use the commands below. Cross-platform builds are published by automation; the [support guide](../SUPPORT.md#compatibility-expectations) explains the current test coverage across platforms.

```sh
prosetta version
prosetta catalog
prosetta probe
```

`catalog` fetches and activates the published catalog and prints its root. Inspect `index.json` in that root for concept IDs; the public CLI has no `list` command. Local authors can use `just list` in a checkout.

A source build embeds the version from `version.json`, so building locally does not automatically produce a `dev` engine. To explicitly use local content with `catalog`, set `PROSETTA_CATALOG` to the absolute path of your `problems/` tree. See [catalog selection](cli.md#catalog-selection).

If a toolchain is missing, install it and run `prosetta probe` again. See [troubleshooting](cli.md#troubleshooting) for other common issues.
