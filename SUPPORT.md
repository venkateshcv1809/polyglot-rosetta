# Support

[Project overview](README.md) · [Troubleshooting](docs/cli.md#troubleshooting) · [Security policy](SECURITY.md)

Polyglot Rosetta provides repository-based, best-effort support. It has no published response-time commitment, paid support plan, or service-level agreement.

## Get help

Start with [getting started](docs/getting-started.md) and the [CLI troubleshooting guide](docs/cli.md#troubleshooting). Check `prosetta <command> --help` for the options supported by your installed binary.

Use repository issues for reproducible bugs, documentation gaps, and feature proposals when issues are available to you. While access is restricted, use the repository's internal issue tracker or an existing private maintainer contact. Do not include secrets or private source code in a public report.

## Report a bug

Include:

- The output of `prosetta version`, your operating system and architecture, and whether you built from source.
- The exact command, working directory context (checkout or standalone install), and expected and actual behavior.
- The full concept ID, language, relevant toolchain version, and a minimal source example.
- Relevant diagnostic output and `--format=json` output when available, including the process exit status.
- Any catalog URL override, whether the content is local or downloaded, and relevant cache configuration with private details removed.

For browser integrations, also include the browser version, WASM asset version, callback involved, and a minimal reproduction without private runtime assets. Distinguish core errors from failures in the host's compiler or worker implementation.

## Compatibility expectations

The pinned tools in [`mise.toml`](mise.toml) define the development environment. CI currently runs on Linux; release cross-compilation produces six OS/architecture combinations, but does not establish that every language harness has been exercised on each platform. Include platform-specific failures in reports.

The project is in the `v0.x` stage. JSON payloads carry a schema version; consumers should reject unsupported versions and handle command failures. There is no published long-term-support schedule or deprecation window. Review release notes and validate your integration before upgrading.

Security-sensitive reports follow [SECURITY.md](SECURITY.md) instead of the ordinary bug-report workflow.
