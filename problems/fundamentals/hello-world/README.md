# Hello World — Solution guide

[Problem statement](PROBLEM.md)

## Overview

Return a greeting to verify the solution interface and evaluate the same behavior across five languages.

## Approach

Read the `name` field and concatenate it with the exact prefix `Hello `, including the trailing space. Preserve the name without trimming, changing case, or adding punctuation.

## Correctness

Every valid input supplies a string named `name`. Concatenating `Hello ` with that string produces exactly the output required by the statement. This also holds when the name is empty or contains whitespace or Unicode text.

## Complexity

Let `N` be the size of the name in the language's string representation. Constructing the result copies the prefix and name, so both the time and returned output space grow linearly. The prefix has constant size.

- **Time Complexity:** $\mathcal{O}(N)$
- **Space Complexity:** $\mathcal{O}(N)$

## Worked example

For `{"name":"World"}`, concatenating `Hello ` and `World` returns `"Hello World"`. An empty name returns `"Hello "`, including its final space.

## Reference Implementation

```go
package main

type Solution struct{}

// Solve returns a greeting for the supplied name.
func (Solution) Solve(input any) any {
	name := input.(map[string]any)["name"].(string)
	return "Hello " + name
}
```

Reference implementations: [Go](go/main.go), [Python](python/main.py), [Rust](rust/main.rs), [TypeScript](typescript/main.ts), and [Zig](zig/main.zig).
