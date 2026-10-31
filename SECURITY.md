# Security policy

[Project overview](README.md) · [Architecture](docs/architecture.md)

## Execution trust model

Polyglot Rosetta evaluates trusted practice code. Native submissions and catalog reference solutions run as child processes with the invoking user's filesystem, environment, and network access. Temporary workspaces, process deadlines, and bounded output do not provide a sandbox, memory isolation, or a security boundary against hostile submissions.

Review source and catalog provenance before running code. A service accepting untrusted submissions needs an independently designed execution isolation layer, resource controls, and credential handling. That layer is outside this repository's native runtime.

Browser hosts own compiler delivery, workers, execution deadlines, and applicable browser security controls. The WASM core validates requests and results; it does not implement those host protections. See the [browser guide](docs/browser-integration.md).

## Catalog and artifact integrity

Native catalog synchronization checks downloaded assets against the SHA-256 values in the selected manifest. The manifest itself is trusted; a checksum does not authenticate a publisher. Select catalog endpoints you trust and protect their publishing credentials. Browser hosts must perform their own manifest and asset verification before passing bytes to the core.

Release automation produces `SHA256SUMS`. Use it to detect mismatches between downloads and the published checksum file. The release workflow does not provide artifact signatures or a separate provenance verification mechanism.

Hidden test cases are distributed with catalog content and can be inspected. Do not put credentials, personal data, or confidential assessment material into exercises, templates, or test cases.

## Report a vulnerability

Use the repository's **Security → Report a vulnerability** option if private vulnerability reporting is enabled and available to you. Otherwise, contact a maintainer through an existing private channel or request a private reporting route without disclosing exploit details. This repository does not publish a dedicated security email address.

Include the affected version or commit, affected component, impact, reproduction steps, and any proposed mitigation. Use a minimal example and remove credentials and personal information. Coordinate disclosure with maintainers before opening a public issue containing vulnerability details.

## Security maintenance

There is no published security response SLA, bug bounty, or supported-version matrix. Reports should identify the exact affected version and, where practical, whether the issue also reproduces on the current default branch. Release notes should identify security fixes and any required consumer action.
