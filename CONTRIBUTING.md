# Contributing to Polyglot Rosetta

[Project overview](README.md) · [Development guide](docs/development.md) · [Exercise authoring](docs/authoring.md)

Contributions should make cross-language learning clearer, evaluation more reliable, or integration easier. While the repository is private, these instructions apply to collaborators with access. Public contribution channels become available only when repository access allows them.

## Propose a change

For a defect, describe the observed behavior, expected behavior, engine version, and smallest reproducing example. Use the [support guide](SUPPORT.md) for the information to include. Report vulnerabilities through the [security process](SECURITY.md).

For a new command, language, catalog format, or integration boundary, open an issue or discuss the proposal with a maintainer before a large implementation. Include the user need, compatibility impact, and how the change will be validated. Documentation fixes and focused bug fixes can go directly to a pull request.

## Work locally

1. Create a branch from the target branch of your intended pull request. Fork only when repository permissions require it.
2. Install the pinned tools and follow the [local setup guide](docs/development.md#local-setup-and-builds).
3. Keep changes focused. Update documentation and examples whenever behavior changes.
4. Run `just check-go` for Go changes and `just test-wasm` for shared contracts or browser bindings. Run `just build-bundle` for catalog or template changes, and evaluate affected reference solutions in every supplied language.
5. Review the diff for credentials, local paths, generated output, and unrelated edits before opening the pull request.

For documentation-only changes, verify local links, check commands against current help and recipes, and execute affected examples where practical. Full runtime tests are unnecessary unless behavior also changes.

## Pull request expectations

Explain the problem, the resulting behavior, and the checks you performed. Call out any skipped checks and why. Provide a before/after example when it makes the change easier to review.

Changes to IDs, JSON contracts, catalog layout, and runtime requirements must explain migration and compatibility implications. Renaming an exercise changes its runtime ID. Do not claim backward compatibility without checking existing callers and fixtures.

Add tests for behavior changes and regressions. Exercise content must meet the [editorial acceptance criteria](docs/authoring.md#editorial-acceptance-criteria). Documentation should use clear language, runnable examples, descriptive links, and claims supported by the implementation.

CI runs on pull requests targeting `develop` and `main`. Follow the active branch policy and maintainer review for your change; this document does not grant merge or release permissions.

## Licensing and collaboration

By submitting a contribution, you agree that it can be distributed under the repository's [MIT License](LICENSE). Submit only material you have the right to contribute and retain required third-party notices. A separate contributor license agreement or developer sign-off is not configured in this repository.

Keep discussion respectful, specific, and focused on the work. Explain technical disagreements with reproducible examples and make review feedback actionable.
