# Security Policy

## Reporting a vulnerability

Please do not report security vulnerabilities in public issues or pull requests.

Use GitHub's private vulnerability reporting or private security advisory feature for
this repository when available. If that feature is unavailable, contact the repository
maintainers privately through GitHub before disclosing details publicly.

## Scope

The library treats input as data and does not execute shell commands. It is not a shell
interpreter and must not be used as a security boundary around arbitrary shell input.
Applications should still validate untrusted variable names, expansion input, and
custom syntax implementations. Custom `Syntax` and `Expander` implementations execute
application code and are responsible for their own resource limits and safety.

Please include a minimal reproducer, the affected version, expected behavior, and actual
behavior in a private report.
