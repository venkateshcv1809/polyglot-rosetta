# CLI and configuration

[Project README](../README.md#documentation) · [Getting started](getting-started.md)

The public binary is `prosetta`. Run `prosetta --help` or `prosetta <command> --help` for built-in usage. Examples below assume the binary is on `PATH`; after a local build, use `./bin/prosetta`.

## Commands

### version

```sh
prosetta version
prosetta version --format=json
```

Prints the version embedded from `version.json`. JSON output includes `schema_version`, `status`, and `version`.

### probe

```sh
prosetta probe
prosetta probe --lang=python --format=json
prosetta probe --clear
```

Inspects the supported language toolchains on `PATH` and replaces the readiness snapshot. `--lang` filters displayed results; inspection still refreshes every language. `--clear` removes the saved snapshot without probing and cannot be combined with a specific language.

A run trusts an existing snapshot. If no snapshot exists, it inspects once. After installing a previously missing toolchain, run `probe` again; a cached unavailable result does not refresh automatically.

### catalog

```sh
prosetta catalog
prosetta catalog --url=https://example.com/catalog/ --format=json
```

Synchronizes a published catalog or resolves a local authoring tree. A remote directory must contain `manifest.json` and the assets it names. Downloads are checked against manifest SHA-256 values, reused by hash, and activated as a full problems tree. JSON success includes `root`.

The manifest host is trusted: hashes check assets against that manifest, not the identity of the publisher. `min_engine` prevents an older release engine from activating incompatible content; `dev` engines bypass that version check.

### run

Replace `CONCEPT_ID` below with a full ID or unique prefix from `just list` or the catalog index.

```sh
prosetta run --id=CONCEPT_ID --lang=go
prosetta run --id=CONCEPT_ID --lang=go --stdin < solution.go
prosetta run --id=CONCEPT_ID --lang=python --code='class Solution:
    def solve(self, data):
        return "Hello " + data["name"]'
prosetta run --id=CONCEPT_ID --lang=go --stdin --all --format=json < solution.go
```

`--id` and `--lang` are required. Languages are `go`, `python`, `rust`, `typescript`, and `zig`. The default source is the reference solution. `--stdin` reads source text, and `--code` supplies inline text; use only one override.

Public cases run by default; `--all` includes hidden cases. There is no public `submit` subcommand; `just submit` is a repository shortcut for `run --all`.

The command exits nonzero on failed evaluation or errors, including with `--format=json`. JSON run status is `passed`, `failed`, or `error`; per-case results distinguish passed, wrong answers, crashes, and timeouts. Integrations should check `schema_version` (currently `1`); result fields are defined in [`internal/api/run/result.go`](../internal/api/run/result.go).

Native evaluation allows 10 seconds per test case and at least 60 seconds per build command. Captured output retains a bounded tail of each stream (1 MiB by default). These are implementation defaults, not public command-line settings or memory limits. `solution_time_ns` measures the solution call; total duration includes additional work and is not a controlled cross-language benchmark.

For automation, read JSON from standard output and diagnostics from standard error, and preserve the exit status. Argument parsing failures can occur before a JSON result is produced; handle nonzero exits with empty or non-JSON output. Do not infer success from the presence of valid JSON alone.

## Catalog selection

For `run`, selection proceeds through:

1. `PROSETTA_CATALOG`, when set, exclusively selects that local tree.
2. Local `problems/` trees relative to the current directory and the binary's project directory.
3. The activated native catalog cache.

Without an explicit override, a concept absent from a local tree can be resolved in the activated cache. Malformed local catalogs produce errors.

For `catalog`, `PROSETTA_CATALOG` selects local content and overrides URL settings. Otherwise `--url` or `PROSETTA_CATALOG_URL` selects a remote directory. Without a URL override, released versions use the [Pages catalog channel](https://venkateshcv1809.github.io/polyglot-rosetta/catalog/); engines whose embedded version is `dev` use local content. A local source build can still have a release version.

```sh
export PROSETTA_CATALOG=/absolute/path/to/polyglot-rosetta/problems
prosetta catalog
```

Concept IDs are `c-` followed by the first 32 hexadecimal digits of SHA-256 over the catalog-relative path, such as `fundamentals/hello-world`. Moving or renaming that path changes its ID. Full IDs are recommended for clients; unique prefixes containing at least eight hex digits work interactively, with or without `c-`.

## Native configuration

On first cache use, the CLI creates `prosetta/config.json` under the OS user configuration directory. `PROSETTA_CONFIG` selects a different file. The default cache root is `prosetta` under the OS user cache directory.

Example configuration (replace the root with your own absolute path):

```json
{
  "schema_version": 1,
  "cache": {
    "root": "/home/user/.cache/prosetta",
    "catalog": "catalog",
    "toolchain": "toolchains.json"
  }
}
```

`root` must be an absolute clean path. `catalog` and `toolchain` must be clean relative paths that do not overlap. Unknown fields and unsupported schema versions are rejected. `PROSETTA_CATALOG_CACHE`, `PROSETTA_TOOLCHAIN_CACHE`, and `catalog --cache` are unsupported.

Temporary evaluation workspaces use the OS temporary directory, including `TMPDIR` where supported. Browser hosts manage their own storage. Native editor integrations should share this CLI configuration.

## Troubleshooting

- **No catalog available:** run `prosetta catalog`, run from the repository root, or set `PROSETTA_CATALOG` to a valid local tree.
- **Concept not found or ambiguous:** use the full ID from the current index or `just list`. Directory names are not run IDs.
- **Toolchain unavailable after installation:** confirm the required executables are on `PATH`, then run `prosetta probe` to replace the snapshot.
- **Reference solution runs instead of your edit:** pass `--stdin < solution.go` or `--code`.
- **Build failure:** check the selected language's [solution interface](authoring.md#solution-interfaces) and the pinned toolchain versions in `mise.toml`.
- **Catalog requires a newer engine:** update the binary or select a compatible catalog deployment.
- **Invalid cache configuration:** inspect the file selected by `PROSETTA_CONFIG` or the OS user config directory and follow the path rules above.
