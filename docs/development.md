# Development, testing, and releases

[Project README](../README.md#documentation) · [Architecture](architecture.md) · [Writing exercises](authoring.md)

See [CONTRIBUTING.md](../CONTRIBUTING.md) for review expectations and the [public release checklist](release-readiness.md) before changing visibility or publishing artifacts.

## Local setup and builds

Use the pinned tools in [`mise.toml`](../mise.toml), or install matching tools yourself. Run commands from the repository root.

```sh
mise trust
mise install --locked
just help
just build-local
```

`just build-cli` writes `bin/prosetta`; `just build-internal` writes `bin/generate`; `just build-wasm` writes `dist/prosetta.wasm` and `dist/wasm_exec.js`. `just build` builds the public CLI and WASM assets, and `just build-local` also builds the maintainer binary. Catalog packaging is separate: `just build-bundle` writes `dist/problems/`.

Without `just`, the native build commands are:

```sh
go build -o bin/prosetta ./cmd/prosetta
go build -o bin/generate ./cmd/generate
```

Generated build output should not be committed. `just clean` removes `bin/`, `dist/`, and coverage reports.

## Checks

```sh
just test        # Go tests
just lint        # go vet
just test-race   # Go tests with the race detector
just test-wasm   # Go WASM tests executed using Node
just check-go    # go vet, race tests, coverage reports
```

Race tests require a supported native platform and C toolchain. `just test-cover` writes `coverage.out` and `coverage.html`. `just format` runs Go formatting; `just tidy` tidies and verifies module dependencies.

The WASM suite compiles Go tests for `js/wasm` and executes them through Go's Node wrapper. It exercises exports, promises, catalog activation, host rebinding, and errors with fixture callbacks. It does not execute browser language compilers or separately load the standalone release WASM artifact.

To use the repository's pre-commit gate:

```sh
git config core.hooksPath .githooks
```

The hook runs `just check-go`.

## Versions

Root [`version.json`](../version.json) is embedded at compile time by `version.go`:

```json
{
  "version": "v0.3.0",
  "min_engine": "v0.1.0"
}
```

`version` is the engine version and must be `dev` or a `v`-prefixed release version. `min_engine` is the minimum engine version placed in generated catalog manifests. Update compatibility requirements deliberately when changing the catalog format or behavior. Building from a checkout preserves this metadata; it does not automatically switch the version to `dev`.

## Publishing

Current behavior is defined by the workflows in [`.github/workflows/`](../.github/workflows/):

- **Quality Checks:** pushes and pull requests targeting `develop` or `main`, merge groups, manual runs, and reusable calls. Runs workflow linting, Go vet, native/WASM builds, catalog packaging, race tests with coverage, and WASM tests. Total Go coverage must be at least 70%.
- **Publish Catalog:** relevant changes on `main` or manual dispatch on `main`, after the quality gate. Builds `dist/problems/*`, copies assets into the Pages source, builds Pages with Jekyll, and deploys the site. The moving catalog channel is `/polyglot-rosetta/catalog/`. Configure repository Pages to use GitHub Actions.
- **Release Binaries:** `v*` tags, after the quality gate. The tagged commit must be in `main` history, and the tag must exactly match the embedded engine version.

Build release artifacts locally with:

```sh
just release-cli
just release
```

`release-cli` cross-compiles Linux (`linux`), macOS (`darwin`), and Windows (`windows`) for `amd64` and `arm64`. Files are named `bin/prosetta-<version>-<os>-<arch>`, with `.exe` on Windows. `release` additionally builds the WASM core and matching loader.

The release workflow publishes those six CLI binaries, `prosetta.wasm`, `wasm_exec.js`, and `SHA256SUMS` to GitHub Releases. Versions containing a prerelease suffix are marked as prereleases. Catalog assets are published separately through Pages; the current workflows do not publish `catalog-v*` releases or attach catalogs to runtime releases.

Consumers should fetch the manifest and its named archives directly rather than wrapping them in another tarball. Applications should reference deployed assets instead of committing generated catalog or toolchain blobs.

Publishing settings and consumer access require separate verification. A successful build does not establish that Releases or Pages are reachable by the intended audience, particularly while the repository is private. The catalog channel changes independently of tagged engine assets; applications that require reproducible content should retain a compatible catalog deployment and its manifest.
