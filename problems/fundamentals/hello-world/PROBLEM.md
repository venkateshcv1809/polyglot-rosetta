# Hello World

## Problem Statement

Given a name, return a greeting formed by placing `Hello ` before it. This introductory exercise verifies the solution interface and your selected language toolchain.

## Input Format

- A JSON object with a string field named `name`.

## Output Format

- A string containing `Hello ` followed by the input name, preserving the name exactly. Return the value through the solution interface; do not print it as the answer.

## Constraints

- The `name` field is required and always contains a string; handling missing fields or other types is outside this exercise.
- The string may be empty or contain whitespace and Unicode characters. Do not trim or normalize it.

## Examples

### Example 1

**Input:**
```json
{"name":"World"}
```

**Output:**
```json
"Hello World"
```

## Explanation

The solution prefixes the supplied name with `Hello `.
