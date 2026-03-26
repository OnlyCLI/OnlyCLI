# Security Policy

## Supported Versions

Security updates are provided for the **latest release** of the OnlyCLI tool. Older releases may not receive patches. Users should upgrade to the current release to benefit from security fixes.

## Reporting a Vulnerability

Please report security vulnerabilities **privately** by email to **security@onlycli.dev**.

Do not open a public GitHub issue for undisclosed security problems.

We aim to acknowledge receipt within **48 hours** and will work with you on understanding and addressing the issue. Response times can vary around holidays or low maintainer availability; if you do not hear back within a reasonable window, you may send a brief follow-up to the same address.

## What to Include in Your Report

To help us assess and fix issues quickly, please include where possible:

- A clear description of the vulnerability and its impact
- Steps to reproduce, or a minimal proof of concept
- Affected component (for example parser, codegen, CLI entrypoint) and version or commit if known
- Any suggested mitigation you are aware of (optional)

## Disclosure Policy

We follow **coordinated disclosure**:

- We will work on a fix and may prepare an advisory or release notes describing the issue after a fix is available.
- Please allow maintainers reasonable time to develop and ship a fix before public disclosure.
- As a guideline, we target resolution and disclosure within approximately **90 days** from report acknowledgment when feasible. Complex issues may take longer; we will communicate status when we can.

## Scope

This policy applies to the **OnlyCLI tool itself** (this repository and official releases): the generator, its templates, and runtime behavior of the `onlycli` CLI as distributed by the project.

**Out of scope**: **Generated CLIs** and other code produced by users with OnlyCLI are the **responsibility of the user** who generated and deployed them. Vulnerabilities in user applications, custom templates, or third-party dependencies of generated projects should be reported to the appropriate owners, not treated as in-scope for this policy unless the flaw is in the generator or bundled official templates shipped with OnlyCLI.

If you are unsure whether something is in scope, email **security@onlycli.dev** and we will triage it.
