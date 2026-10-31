# Public release readiness

[Project overview](../README.md) · [Release workflow](development.md#publishing)

Use this checklist before changing repository visibility or publishing a release. These are verification tasks, not a statement that repository settings, access controls, or past history have already been audited. Publishing the repository and publishing its runtime or catalog are separate actions.

## Before making the repository public

- [ ] Review the complete Git history, branches, tags, issue content, attachments, and workflow logs for credentials, private data, internal URLs, and material that should not become public. Rotate exposed credentials; removing a current file does not remove its history.
- [ ] Confirm ownership and redistribution rights for code, exercises, documentation, and any third-party material. Keep required notices with distributed content.
- [ ] Confirm the project description, catalog scope, supported integrations, and limitations in the README match the release candidate.
- [ ] Enable and verify a private vulnerability reporting route, or publish a working private security contact in `SECURITY.md`.
- [ ] Configure repository issues and contributor access, then verify the routes described in `SUPPORT.md` and `CONTRIBUTING.md`.
- [ ] Review Actions permissions, fork pull request behavior, deployment environments, required checks, and protection for the default and release branches.
- [ ] Check links and examples as a reader with the intended access level. Confirm that source, Releases, and Pages URLs resolve after the visibility change.
- [ ] Check Pages visibility and deployment access separately from source visibility. Review catalog contents, including hidden cases, for public distribution.

## Before tagging an engine release

- [ ] Update `version.json` deliberately, including `min_engine` when catalog compatibility changes.
- [ ] Pass the quality workflow on the release candidate. Confirm any changed language harness against the pinned toolchain and evaluate affected reference solutions.
- [ ] Smoke-test the getting-started commands from a clean checkout and the standalone release workflow with an isolated configuration and cache.
- [ ] Test claimed platform behavior on the relevant platform; cross-compilation alone verifies a binary can be built.
- [ ] Validate browser integrations against the candidate WASM core and matching loader when the browser contract changes. Unit tests with callback fixtures do not replace host integration testing.
- [ ] Write release notes covering user-visible changes, compatibility, migration steps, known limitations, and security fixes where relevant.
- [ ] Confirm the release commit is in `main` history and the proposed tag exactly matches the embedded version. Tagging `v*` triggers publication under the current workflow.

## After publishing

- [ ] Download the published artifacts, verify `SHA256SUMS`, and smoke-test `version`, catalog activation, toolchain probing, and a reference run.
- [ ] Confirm the catalog manifest and every named asset are accessible from the consumer environment. The Pages catalog is a moving channel; retain a known compatible deployment if an integration requires a fixed catalog.
- [ ] Check that the release notes, source tag, README, and available assets describe the same release.
- [ ] Record remaining platform or integration limitations and route follow-up work through repository issues.

The [development guide](development.md#publishing) describes what automation enforces. Repository settings, historical content review, distribution rights, and consumer smoke tests require separate verification.
