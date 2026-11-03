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

Scaffolding registers the new item in its parent and copies templates into the new directory. It refuses existing target directories. Generated reference files contain placeholder implementations and sample test data; complete them before considering the exercise ready.

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

For Rust and Zig, return plain text for string expectations and valid JSON for other values; JSON whitespace and object key order do not affect comparison. Use standard-library implementations and the harness's provided types; do not assume third-party dependencies are installed. The harness handles invoking the solution and reporting results.

The scaffold copies shared templates to `<language>/main.<extension>` as the starting point for reference implementations. Practice starters are separate: they use `templates/concept/<language>/template.<extension>` unless a concept supplies `<language>/template.<extension>.txt`. For example, a concept-specific Go starter belongs in `go/template.go.txt`.

## Input, output, and solution lifecycle

The shared interface is a way to drive your implementation. Define your own classes, structs, functions, and data structures in the solution file, then use `Solve`/`solve` to construct them, perform the requested work, and return observable results. The harness supplies the entry point; do not read testcase data from standard input or add a competing `main` function.

The input is a JSON object. Its fields can contain nested objects, arrays, strings, finite numbers, booleans, and null. Define the meaning of those fields in the problem statement. JSON does not directly represent pointers, shared object identity, cyclic references, or executable callbacks. Encode relationships with IDs, adjacency lists, or another documented representation, and construct the corresponding objects inside the solution.

Native evaluation starts a new process for every testcase and invokes the solution once in that process. State created inside that call is available for its full duration. State is not shared between testcases. Browser hosts must launch each requested case in a fresh runtime environment to preserve this behavior.

Return the answer; printing or logging is only for debugging and is subject to captured-output limits. Go, Python, and TypeScript return JSON-compatible values. Rust and Zig retain their text-return interfaces: an expected string `"hello"` requires the returned text `hello`, while an expected array `[1,2]` requires JSON text such as `[1,2]`. For a string containing quotes, newlines, or backslashes, return the actual string characters rather than a JSON string literal.

Language-specific details:

- **Go:** input objects are `map[string]any`, arrays are `[]any`, and numbers decode as `float64`; a JSON integer does not arrive as a Go `int`. Other values are `string`, `bool`, and `nil`. Returned values must be accepted by `encoding/json`, including exported fields when returning structs.
- **Python:** inputs use `dict`, `list`, `str`, `int`/`float`, `bool`, and `None`. The starter uses `typing.Any`. Return values must be serializable by the standard JSON encoder; non-finite numbers are rejected.
- **TypeScript:** narrow `unknown` to the documented input shape. Use plain objects, arrays, strings, finite numbers, booleans, or null as results. Execution is synchronous; Promises, `undefined`, array holes, `bigint`, and class instances are not supported return values. Convert internal classes and collections to plain data before returning.
- **Rust:** `JsonValue` has `Null`, `Bool(bool)`, `Number(f64)`, `String(String)`, `Array(Vec<JsonValue>)`, and `Object(BTreeMap<String, JsonValue>)` variants. Match the required variants before accessing fields. The harness supplies this type; it does not supply `serde_json` or another dependency.
- **Zig:** inspect the active `std.json.Value` tag before accessing fields such as `.object`, `.array`, or `.string`. Numbers can use `.integer` or `.float`. Use the supplied allocator for dynamically constructed output. Return a slice that remains valid after the function returns; a local stack array or already-freed allocation is invalid. The native harness releases its arena after reporting the result.

Catalog JSON numbers currently decode through Go's `float64` representation, and Rust inputs also use `f64`. Keep integer data within the interoperable safe range of `-(2^53 - 1)` through `2^53 - 1`; use documented decimal strings when exact larger integers are part of the exercise. Floating-point answers have no approximate matcher, so choose a deterministic representation or explicitly defined rounding.

## Stateful DSA and design-pattern exercises

An exercise can define a sequence of operations inside one input object. For example, a stack problem could use:

```json
{
  "initial": [],
  "operations": [
    { "action": "push", "value": 3 },
    { "action": "pop" }
  ]
}
```

If the statement says only `pop` produces output, the expected value is `[3]`. The solution constructs a stack and executes the sequence within one call. Specify behavior for empty pops and other invalid operations in the problem rather than leaving it to the runner.

For an Observer-pattern exercise, an input could describe subscriptions and events. The solution constructs subjects and observers, processes the events, and returns the notification history. That history tests externally visible behavior; matching it does not prove the solution uses the intended design pattern. Review the implementation and explain the participant roles in the solution guide.

These are exercise-defined data formats. The runner does not interpret operation names or prescribe a universal operation schema. Graph traversal, cache behavior, factory selection, and other concepts can each define the data that makes their behavior observable.

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

Use sequential numeric test IDs starting at `1`, with the array entries in the same order. Execution follows array order; the runner does not sort by ID. Numeric IDs are normalized to strings in runtime results; legacy string IDs remain readable. The `input` must be a JSON object; `expected` can be any JSON-compatible value. Include useful edge cases, public examples, and hidden cases for full evaluation. Hidden cases are distributed as catalog content and remain inspectable.

Native and WASM evaluation use one shared Go comparison rule across languages: strings match exactly, object key order is ignored, array order and length matter, and booleans are distinct from numbers. Equivalent JSON numbers such as `1` and `1.0` compare equally. Solutions must return JSON-compatible values; TypeScript array holes and `undefined` values are rejected. Both hosts return raw harness output for Go to parse and compare; browser hosts do not grade answers.

There are no custom matchers, unordered-array comparisons, floating-point tolerances, or checks of object identity or internal implementation structure. If several output orders would be valid, define a canonical order in the statement and have the solution return it. A crash or uncaught error is an execution failure, not a way to satisfy an expected error; return a documented error value when that is part of the exercise.

The shared [testcase template](../templates/concept/testcases.json) demonstrates the supported data shapes with numbered examples: strings, integers, decimals, booleans, null, arrays, matrices, nested objects, mixed arrays, empty values, and Unicode or escaped text. Every example reads `input.value` and expects that same value back; the final example demonstrates a hidden case. This is a catalogue of input/output shapes, not a requirement that exercises implement an echo operation. Replace the samples with cases for your concept and update the problem examples to match. The minimal starter implementations are incomplete and are not required to pass the sample cases.

Validate every reference solution against both public and hidden cases. The [Hello World exercise](../problems/fundamentals/hello-world/) demonstrates the complete layout.

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

## Updating shared starters

Editing a shared starter changes future scaffolds and the practice starter shown for concepts that use the shared fallback. It does not overwrite existing `<language>/main.<extension>` reference solutions, concept-specific starter overrides, or previously copied practice files. Rebuild and republish the catalog bundle to distribute shared starter changes.

The internal [harness templates](../internal/execution/harnesses/) are separate from learner-facing starters. They are embedded in the engine; changing them requires an engine rebuild. Starter comments and authoring guidance do not change the solution interface or catalog schema.
