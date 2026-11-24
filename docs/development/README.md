# Development notes

Use the repository Makefile as the canonical local command entry point.
`make check` runs formatting checks, tests, the race detector, vet, and
golangci-lint. `make install-hooks` installs optional local commit hooks.
