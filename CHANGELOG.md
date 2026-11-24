# Changelog

Notable changes are documented here. This project follows semantic versioning for tagged
releases.

## Unreleased

- Generalize the expansion engine so syntaxes and kinds are application-extensible.
- Make `Expression` syntax-neutral; dialect-specific parse data is carried in `Expression.Data`.
- Organize the Bash parser, strategies, environment adapters, and pattern matcher into
  cohesive `internal/` packages.
- Add runnable custom-syntax examples, architecture documentation, and repository
  contribution, security, governance, ownership, lint, CI, and hook tooling.
