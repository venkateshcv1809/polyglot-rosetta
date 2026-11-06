---
layout: release
title: Polyglot Rosetta
description: Compare coding exercises in Go, Python, Rust, TypeScript, and Zig.
---

# Learn a concept. Compare five implementations.

Polyglot Rosetta combines exercises, explanations, reference solutions, and evaluation in one workflow. Work through a problem in **Go, Python, Rust, TypeScript, or Zig**, then compare how each language expresses the same idea.

## Explore the project

- [**Problem catalog**]({{ '/catalog/' | relative_url }}) - Browse exercises, explanations, reference solutions, and test cases.
- [**Releases**](https://github.com/venkateshcv1809/polyglot-rosetta/releases) - Download the CLI and WebAssembly core.
- [**Source code**](https://github.com/venkateshcv1809/polyglot-rosetta) - Explore the engine, catalog tools, and exercise content.

## Each exercise has a full learning loop

Start with a problem statement and its constraints. Read an explanation with examples and complexity analysis, then study or write an implementation. Public cases give quick feedback; hidden cases check the broader solution.

## Use the engine in your workflow

- **Practice locally:** the `prosetta` CLI builds and evaluates solutions with your installed language toolchains.
- **Build an integration:** the shared Go core can run as WebAssembly, with its host providing compilation and execution capabilities.
- **Author or distribute exercises:** keep problem content in a catalog that can be packaged and updated independently of engine releases.

## Get started

Install the Go version listed in `go.mod` and [just](https://just.systems/), then build the CLI and run the sample exercise:

```sh
just build-cli
./bin/prosetta run --id=c-2b23ca0786e0eb237f64cf2b8759bcc3 --lang=go
```

See the [getting started guide](https://github.com/venkateshcv1809/polyglot-rosetta/blob/main/docs/getting-started.md) for setup and the complete workflow. The [CLI reference](https://github.com/venkateshcv1809/polyglot-rosetta/blob/main/docs/cli.md) covers commands and options.

Native evaluation runs with your account's permissions. Use it for trusted practice code; see the [security policy](https://github.com/venkateshcv1809/polyglot-rosetta/blob/main/SECURITY.md) before building a service around it.
