# Writing exercises

[Project README](../README.md#documentation) · [Development guide](development.md)

Run maintainer commands from the repository root. They use the local `problems/` and `templates/` trees. `generate` is a separate maintainer binary; its commands are not exposed by the public CLI or WASM core.

## Exercise content

Each exercise lives under `problems/<category>/<concept>/` and contains:

- `PROBLEM.md`: statement, input/output format, constraints, and examples.
- `README.md`: approach and complexity.
- `testcases.json`: public and hidden cases.
- `<language>/main.<extension>`: single-file reference solutions.
- `info.json`: title, description, directory ID, and difficulty (`Easy`, `Medium`, or `Hard`).

Category `info.json` files register children and their order. Scaffolding maintains these references for you. Use lowercase kebab-case directory names; moving an exercise changes its runtime ID.

## Scaffold content

```sh
just category
just concept
```

These commands prompt for missing fields. Category parents are optional (empty means root); a concept needs an existing parent category. You can prefill fields:

```sh
just category --id=algorithms --title=Algorithms --description="Core algorithms"
just concept --parent=algorithms --id=two-sum --title="Two Sum" --difficulty=Easy --description="Find a pair with a target sum"
```

Scaffolding registers the new item in its parent and copies templates into the new directory. It refuses existing target directories. Generated reference files contain TODO implementations and sample test data; complete them before considering the exercise ready.

To use the standalone maintainer binary:

```sh
just build-internal
./bin/generate --help
./bin/generate list
./bin/generate list --search=greeting
```

## Solution interfaces

Solutions are single files evaluated by a generated harness. Implement the entry point from the matching [language template](../templates/concept/):

- Go: `func (Solution) Solve(inputData any) any` in `package main`.
- Python: `Solution.solve(self, input_data)` returning a value.
- TypeScript: exported `Solution` class with `solve(inputData: unknown): unknown`.
- Rust: `Solution::solve(input_data: &JsonValue) -> String`, using the harness-provided `crate::JsonValue`.
- Zig: `Solution.solve(allocator: std.mem.Allocator, input_data: std.json.Value) ![]const u8`.

For Rust and Zig, return plain text for string expectations and compact JSON for structured values. Use standard-library implementations and the harness's provided types; do not assume third-party dependencies are installed. The harness handles invoking the solution and reporting results.

The scaffold copies shared templates to `<language>/main.<extension>` as the starting point for reference implementations. Practice starters are separate: they use `templates/concept/<language>/template.<extension>` unless a concept supplies `<language>/template.<extension>.txt`. For example, a concept-specific Go starter belongs in `go/template.go.txt`.

## Test cases

`testcases.json` is an array:

```json
[
  {
    "id": 1,
    "description": "Greet a name",
    "input": { "name": "World" },
    "expected": "Hello World",
    "hidden": false
  }
]
```

Use distinct, non-empty test IDs; strings are recommended, and numeric IDs are also accepted and normalized to strings. The `input` must be a JSON object; `expected` can be any JSON-compatible value. Include useful edge cases, public examples, and hidden cases for full evaluation. Hidden cases are distributed as catalog content and remain inspectable.

String expectations compare as text; structured outputs compare semantically. Validate every reference solution against both public and hidden cases. The [Hello World exercise](../problems/fundamentals/hello-world/) demonstrates the complete layout.

```sh
just list
# Replace CONCEPT_ID with the new exercise's full ID.
just submit --id=CONCEPT_ID --lang=go
```

Repeat for each supplied language. `just list` reads live metadata; it does not need a rebuilt bundle.

## Editorial acceptance criteria

Before submitting an exercise:

- State one learning objective and define every input field, output value, and precondition.
- Specify behavior for empty inputs, duplicates, ordering, Unicode, and numeric limits when relevant. Keep numbers within the ranges every supplied implementation can represent reliably.
- Make examples executable test cases with exactly matching expected values.
- Explain the algorithm, why it works, and the meaning of variables used in complexity bounds. Count returned output space explicitly.
- Supply an idiomatic reference implementation for each advertised language. Document any language-specific tradeoffs.
- Check public and hidden cases in each supplied language, and confirm that a plausible incorrect implementation fails an appropriate case.
- Remove scaffold placeholders and keep starter files free of the completed solution.
- Use original or appropriately licensed text and examples; record attribution when adapting material.

Catalog packaging checks structure and content loading; it is not an editorial review or proof that the reference implementations are correct.

## Package the catalog

```sh
just build-bundle
```

This writes the index, shared templates, category archives, and their hashed manifest to `dist/problems/`. The minimum compatible engine comes from [`version.json`](../version.json).

Packaging does not run reference solutions; evaluate them first. See [publishing](development.md#publishing) for deployment.
