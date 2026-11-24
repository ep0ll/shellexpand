# Governance

## Project ownership

This repository is maintained by the owners and approvers listed in [OWNERS](OWNERS).
GitHub review requests use the root-level [CODEOWNERS](CODEOWNERS) file. The
Kubernetes-style [OWNERS_ALIASES](OWNERS_ALIASES) file is available to tools
that understand that convention.

## Decision making

Routine bug fixes, tests, documentation, and internal refactors may be merged
by maintainers after review and successful checks. Public API changes,
syntax compatibility changes, and changes to the extension contracts require
maintainer review and should be discussed in an issue or pull request before
implementation.

Maintainers aim for consensus. When consensus is not possible, the project
maintainer makes the final decision and records the rationale in the relevant
discussion.

## Releases and compatibility

The project follows semantic versioning once tagged releases are published.
Changes to public Go API and built-in syntax behavior must be documented in
CHANGELOG.md. Compatibility boundaries and intentional deviations from Bash
are documented in the README and research notes.

## Security and conduct

Security reports must follow SECURITY.md. Community participation is governed
by CODE_OF_CONDUCT.md.
