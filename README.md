# shellexpand

[![CI](https://github.com/ep0ll/shellexpand/actions/workflows/ci.yml/badge.svg)](https://github.com/ep0ll/shellexpand/actions/workflows/ci.yml)
[![CodeQL](https://github.com/ep0ll/shellexpand/actions/workflows/codeql.yml/badge.svg)](https://github.com/ep0ll/shellexpand/actions/workflows/codeql.yml)

A composable Go expansion engine with a built-in Bash parameter-expansion dialect. The engine is syntax-agnostic: applications can register their own token syntaxes, define arbitrary expansion kinds, and provide custom expanders without depending on shell concepts.

The core design rule is **recognize a complete token, then dispatch by its exact kind**. Expanders are not tried in a fallback chain. A handler for `${NAME}` therefore cannot accidentally consume `${NAME:-fallback}`.

The library evaluates string expressions; it is not a shell interpreter. It never executes commands or evaluates arbitrary shell source.

## Install

```sh
go get github.com/ep0ll/shellexpand
```

The module targets Go 1.23 and has no third-party runtime dependencies.

## Quick start

The package-level `Expand` function uses all built-in Bash expanders:

```go
package main

import (
    "fmt"

    "github.com/ep0ll/shellexpand"
)

func main() {
    env := shellexpand.NewMapEnv(map[string]string{
        "IMAGE": "alpine",
        "TAG":   "3.21",
    })

    got, err := shellexpand.Expand("FROM ${IMAGE}:${TAG:-latest}", env)
    if err != nil {
        panic(err)
    }
    fmt.Println(got) // FROM alpine:3.21
}
```

Use `ProcessEnv{}` to explicitly adapt the current process environment. Assignment forms such as `${NAME:=word}` call `os.Setenv` and mutate the process environment. Use `NewMapEnv` for isolated, concurrency-safe state.

## Build a selective Bash composer

Every built-in expander is an independent strategy. Register only the operations your application permits:

```go
composer := shellexpand.MustNew(
    shellexpand.DollarVariableExpander(),
    shellexpand.BracedVariableExpander(),
)

env := shellexpand.NewMapEnv(map[string]string{"NAME": "world", "EMPTY": ""})
got, err := composer.Expand(
    "$NAME ${NAME} ${EMPTY-default} ${EMPTY:-fallback}",
    env,
)
if err != nil {
    panic(err)
}
fmt.Println(got)
// world world ${EMPTY-default} ${EMPTY:-fallback}
```

Unregistered expressions are preserved verbatim by default. Add `WithUnselectedError()` to return a typed error instead. `DefaultExpanders()` returns a new slice containing every built-in Bash expander; `DefaultComposer()` returns the immutable shared composer with all of them registered.

## Define a completely different syntax

The generic engine does not assume a dollar sign, braces, variable names, or even variable expansion. A syntax recognizes a complete token and returns a `Kind` plus opaque syntax-specific `Data`. An expander claims that exact kind.

This small example recognizes `{{NAME}}`:

```go
type placeholder struct{ name string }
const placeholderKind shellexpand.Kind = "example.double_brace"

syntax := shellexpand.SyntaxFunc(func(input string, offset int, _ shellexpand.ParseOptions) (shellexpand.SyntaxMatch, bool, error) {
    if !strings.HasPrefix(input[offset:], "{{") {
        return shellexpand.SyntaxMatch{}, false, nil
    }
    close := strings.Index(input[offset+2:], "}}")
    if close < 0 {
        return shellexpand.SyntaxMatch{}, false, fmt.Errorf("unterminated placeholder")
    }
    end := offset + 2 + close + 2
    expr := shellexpand.Expression{
        Kind: placeholderKind,
        Data: placeholder{name: input[offset+2 : end-2]},
    }
    return shellexpand.SyntaxMatch{End: end, Expression: &expr}, true, nil
})

expander := shellexpand.ExpanderFunc{
    ExpansionKind: placeholderKind,
    Func: func(ctx *shellexpand.Context, expr shellexpand.Expression) (string, error) {
        token, ok := expr.Data.(placeholder)
        if !ok {
            return "", fmt.Errorf("unexpected payload %T", expr.Data)
        }
        return ctx.Value(token.name)
    },
}

composer := shellexpand.MustNewWithSyntax([]shellexpand.Syntax{syntax}, expander)
got, err := composer.Expand("image={{IMAGE}}",
    shellexpand.NewMapEnv(map[string]string{"IMAGE": "alpine"}))
```

See the complete runnable example in [examples/customsyntax](examples/customsyntax). Multiple syntaxes can be composed; their slice order defines precedence. Applications may define any non-empty `Kind` value except the reserved `KindUnknown`.

### Extension contracts

- `Syntax.Match` is called at each input byte offset. A successful match must consume a non-empty, in-range byte span and must not consume anything when it returns `false`.
- `SyntaxMatch.End` is an exclusive byte offset. An `Expression` represents the complete source token; the engine verifies that its optional `Raw` value matches the source.
- `Expression.Data` is opaque to the engine and carries dialect-specific parsed data.
- `Expander.Kind()` declares exactly one kind. Duplicate registrations and empty kinds are rejected at composer construction.
- `Context` provides `Lookup`, `Value`, `Assign`, `Names`, and `ExpandWord`. Recursive expansion uses the same composer, environment, and per-call options.
- Syntax and expander implementations should be immutable or safe for concurrent use. Per-expansion state belongs in `Context` or `Expression`.

See [the architecture guide](docs/architecture.md) for package boundaries and the full execution model.

## Built-in Bash dialect

`BashSyntax()` recognizes the supported parameter-expansion subset. Each operation below has its own public factory:

| Form | Expander factory |
|---|---|
| `$NAME`, `${NAME}` | `DollarVariableExpander`, `BracedVariableExpander` |
| `${NAME-word}`, `${NAME:-word}` | `DefaultIfUnsetExpander`, `DefaultIfUnsetOrEmptyExpander` |
| `${NAME=word}`, `${NAME:=word}` | `AssignIfUnsetExpander`, `AssignIfUnsetOrEmptyExpander` |
| `${NAME?word}`, `${NAME:?word}` | `ErrorIfUnsetExpander`, `ErrorIfUnsetOrEmptyExpander` |
| `${NAME+word}`, `${NAME:+word}` | `AlternateIfSetExpander`, `AlternateIfSetAndNonEmptyExpander` |
| `${#NAME}`, `${NAME:offset[:length]}` | `LengthExpander`, `SubstringExpander` |
| `${NAME#pattern}`, `${NAME##pattern}`, `${NAME%pattern}`, `${NAME%%pattern}` | `RemoveShortestPrefixExpander`, `RemoveLongestPrefixExpander`, `RemoveShortestSuffixExpander`, `RemoveLongestSuffixExpander` |
| `${NAME/pattern/repl}`, `${NAME//pattern/repl}`, `${NAME/#pattern/repl}`, `${NAME/%pattern/repl}` | `ReplaceFirstExpander`, `ReplaceAllExpander`, `ReplacePrefixExpander`, `ReplaceSuffixExpander` |
| `${NAME^pattern}`, `${NAME^^pattern}`, `${NAME,pattern}`, `${NAME,,pattern}` | `UppercaseFirstExpander`, `UppercaseAllExpander`, `LowercaseFirstExpander`, `LowercaseAllExpander` |
| `${!NAME}`, `${!prefix*}`, `${!prefix@}` | `IndirectVariableExpander`, `IndirectPrefixNamesStarExpander`, `IndirectPrefixNamesAtExpander` |

The colon variants of default, assignment, error, and alternate operators additionally treat an empty value as unset. Default/alternate words are expanded lazily, only when their branch is selected.

### Compatibility boundaries

This is a parameter-value expander, not a Bash interpreter. It does not implement command or arithmetic substitution, positional/special parameters, arrays, shell functions, quotes, word splitting, pathname expansion, tilde expansion, or general shell parsing. Variable names use the ASCII identifier form. The glob matcher supports `*`, `?`, bracket classes/ranges and escaping; it does not implement locale-dependent matching, POSIX character classes, or extended globs. Substring offsets and lengths count Unicode code points. Prefix-name indirection returns a single space-delimited string.

Unsupported syntax is preserved literally unless an explicit strict option applies. The full compatibility notes are maintained in [docs/research/implementation-notes.md](docs/research/implementation-notes.md).

## Environments

`Env.Lookup(name)` returns both a value and a set flag, distinguishing an unset name from one set to an empty string. Custom environments can implement:

- `MutableEnv` to support assignment expressions.
- `EnumerableEnv` to support prefix-name indirection.

`LookupFunc` adapts an `os.LookupEnv`-style function. `MapEnv` is a concurrency-safe map; `ProcessEnv{}` adapts `os.LookupEnv`, `os.Setenv`, and `os.Environ`.

## Options and errors

- `WithStrictUnset()` errors when an ordinary value lookup is unset.
- `WithMaxDepth(n)` sets the positive recursion limit (default 1000).
- `WithUnsupportedError()` errors when a syntax reports the reserved unknown kind.
- `WithUnselectedError()` errors when a recognized kind has no registered expander.

Errors are designed to work with `errors.Is` and `errors.As`. See the exported error types and sentinel values in the package documentation.

## Development

```sh
make check        # format check, tests, race detector, vet and lint
make install-hooks
go test ./...
go test -race ./...
go vet ./...
```

The repository uses conventional commit subjects. Git hooks are opt-in and can be installed with `make install-hooks`.

## Project documentation

- [Architecture](docs/architecture.md)
- [Implementation research notes](docs/research/implementation-notes.md)
- [Contributing](CONTRIBUTING.md)
- [Security policy](SECURITY.md)
- [Governance](GOVERNANCE.md)
- [Support](SUPPORT.md)
- [Changelog](CHANGELOG.md)

Licensed under the [MIT License](LICENSE).
