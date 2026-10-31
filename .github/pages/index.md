---
layout: default
title: Polyglot Rosetta
---

# Polyglot Rosetta

**Learn a concept once. Implement it across five languages.**

Polyglot Rosetta combines programming exercises, reference implementations, and local evaluation for Go, Python, Rust, TypeScript, and Zig. Use the `prosetta` CLI for practice and editor workflows, or integrate the shared WebAssembly core with your own browser runtimes.

The project is in the `v0.x` stage. The initial Hello World exercise demonstrates the full five-language workflow. Native execution is intended for trusted code and does not sandbox submissions.

## Start here

- [Project overview and quick start](https://github.com/venkateshcv1809/polyglot-rosetta#readme)
- [Getting started](https://github.com/venkateshcv1809/polyglot-rosetta/blob/main/docs/getting-started.md)
- [CLI reference](https://github.com/venkateshcv1809/polyglot-rosetta/blob/main/docs/cli.md)
- [Browser host integration](https://github.com/venkateshcv1809/polyglot-rosetta/blob/main/docs/browser-integration.md)

Source and release links require repository access while the project is private.

## Catalog

This is the moving catalog channel, published independently of engine releases. Clients start with `manifest.json` and fetch the assets it names. The manifest records asset sizes, SHA-256 digests, and the minimum compatible engine. Browser hosts verify these before activation; the native `catalog` command handles synchronization.

<ul>
{% for file in site.static_files %}
{% if file.path contains '/catalog/' %}
  <li><a href="{{ file.path | relative_url | escape }}">{{ file.name | escape }}</a></li>
{% endif %}
{% endfor %}
</ul>

## Downloads and source

- [Releases](https://github.com/venkateshcv1809/polyglot-rosetta/releases): CLI binaries for Linux, macOS, and Windows on amd64 and arm64, plus the WebAssembly core and JavaScript runtime bridge.
- [Source code](https://github.com/venkateshcv1809/polyglot-rosetta)
- [Support](https://github.com/venkateshcv1809/polyglot-rosetta/blob/main/SUPPORT.md)
- [Security policy](https://github.com/venkateshcv1809/polyglot-rosetta/blob/main/SECURITY.md)
- [MIT License](https://github.com/venkateshcv1809/polyglot-rosetta/blob/main/LICENSE)
