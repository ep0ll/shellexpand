# shellexpand

A small, composable Go implementation of a deliberately scoped subset of Bash parameter expansion.

The central design rule is **parse first, dispatch by exact syntax, never fall through**. Each expression is classified into one `Kind`, and each expander handles exactly one `Kind`. In particular, the `${NAME}` expander cannot consume `${NAME-default}` or `${NAME:-default}`. When a composer does not contain the matching expander, that complete original expression is preserved unchanged (or can produce an error).

The project is a Go library, not a shell interpreter. It does not execute commands or evaluate arbitrary shell source.

## Quick start

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

	got, err := shellexpand.Expand(
		"FROM ${IMAGE}:${TAG:-latest}",
		env,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(got) // FROM alpine:3.21
}
```

For read-only access to the process environment, adapt the standard library lookup directly:

```go
env := shellexpand.LookupFunc(os.LookupEnv)
got, err := shellexpand.Expand(input, env)
```

When you also need assignment expansions (`${NAME=word}` / `${NAME:=word}`) or prefix-name indirection (`${!PREFIX*}` / `${!PREFIX@}`), use the zero-value `ProcessEnv{}` adapter. It delegates to `os.LookupEnv`, `os.Setenv`, and `os.Environ`:

```go
got, err := shellexpand.Expand(input, shellexpand.ProcessEnv{})
```

Assignment expansions with `ProcessEnv{}` modify the current process environment. Use `NewMapEnv` when changes should remain local to one environment instance.

## Select only the expansions you allow

Every public expander factory is independent. `$NAME` and `${NAME}` have separate expanders, as do each of the default/alternate/assignment/error operators, each pattern operator, and each case operator.

```go
composer := shellexpand.MustNew(
	shellexpand.DollarVariableExpander(),
	shellexpand.BracedVariableExpander(),
)

env := shellexpand.NewMapEnv(map[string]string{
	"NAME":  "world",
	"EMPTY": "",
})

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

The last two expressions are not partially consumed. If only the plain-variable expanders are registered, the entire default-expression spelling survives untouched.

To make missing registrations explicit, pass `WithUnselectedError()`:

```go
_, err := composer.Expand("${NAME:-fallback}", env, shellexpand.WithUnselectedError())
var selectionErr *shellexpand.UnselectedExpanderError
if errors.As(err, &selectionErr) {
	fmt.Println(selectionErr.Kind) // default_if_unset_or_empty
}
```

`DefaultExpanders()` returns a fresh slice of all built-in expanders; `DefaultComposer()` is the shared immutable composer with all built-ins registered. `New` rejects duplicate kinds and nil expanders, so dispatch is deterministic and does not depend on expander ordering.

## Supported parameter expansion forms

| Syntax | Kind / expander factory |
|---|---|
| `$NAME` | `DollarVariableExpander` |
| `${NAME}` | `BracedVariableExpander` |
| `${NAME:-word}` | `DefaultIfUnsetOrEmptyExpander` |
| `${NAME-word}` | `DefaultIfUnsetExpander` |
| `${NAME:=word}` | `AssignIfUnsetOrEmptyExpander` |
| `${NAME=word}` | `AssignIfUnsetExpander` |
| `${NAME:?word}` | `ErrorIfUnsetOrEmptyExpander` |
| `${NAME?word}` | `ErrorIfUnsetExpander` |
| `${NAME:+word}` | `AlternateIfSetAndNonEmptyExpander` |
| `${NAME+word}` | `AlternateIfSetExpander` |
| `${#NAME}` | `LengthExpander` |
| `${NAME:offset}` / `${NAME:offset:length}` | `SubstringExpander` |
| `${NAME#pattern}` | `RemoveShortestPrefixExpander` |
| `${NAME##pattern}` | `RemoveLongestPrefixExpander` |
| `${NAME%pattern}` | `RemoveShortestSuffixExpander` |
| `${NAME%%pattern}` | `RemoveLongestSuffixExpander` |
| `${NAME/pattern/replacement}` | `ReplaceFirstExpander` |
| `${NAME//pattern/replacement}` | `ReplaceAllExpander` |
| `${NAME/#pattern/replacement}` | `ReplacePrefixExpander` |
| `${NAME/%pattern/replacement}` | `ReplaceSuffixExpander` |
| `${NAME^pattern}` | `UppercaseFirstExpander` |
| `${NAME^^pattern}` | `UppercaseAllExpander` |
| `${NAME,pattern}` | `LowercaseFirstExpander` |
| `${NAME,,pattern}` | `LowercaseAllExpander` |
| `${!NAME}` | `IndirectVariableExpander` |
| `${!prefix*}` | `IndirectPrefixNamesStarExpander` |
| `${!prefix@}` | `IndirectPrefixNamesAtExpander` |

A blank case-modification pattern means “any one character”, matching Bash's default pattern. Substring offsets and lengths are decimal integers (arithmetic expressions are not evaluated). An offset may be negative using either `${NAME: -2}` or the parenthesized literal form `${NAME:(-2)}`. A negative substring length stops that many characters before the end; if that computed end precedes an in-range start, `ErrInvalidSubstringRange` is returned.

## Environments

The environment is injected; expanders do not read process-global state directly. This keeps evaluation deterministic and makes isolated testing straightforward. `ProcessEnv{}` is the opt-in adapter for the current process environment.

```go
type Env interface {
    Lookup(name string) (value string, set bool)
}
```

The boolean is important: it distinguishes an unset variable from a variable explicitly set to `""`. `LookupFunc` adapts functions such as `os.LookupEnv`. `ProcessEnv{}` adapts `os.LookupEnv`, `os.Setenv`, and `os.Environ` to all three environment interfaces. `MapEnv` is a concurrency-safe, mutable, enumerable map implementation:

- `NewMapEnv` copies the input map.
- `Set` supports `${NAME=word}` and `${NAME:=word}`.
- `Names` supports `${!prefix*}` and `${!prefix@}`; results are sorted for deterministic output.
- `NewListEnv` accepts `NAME=value` strings, splitting only on the first `=`.

Custom environments can opt into assignment by implementing `MutableEnv`, and prefix enumeration by implementing `EnumerableEnv`. An assignment expansion against a read-only environment returns `ErrReadOnlyEnvironment`. Prefix indirection against an environment without `EnumerableEnv` returns `ErrVariableNameUnavailable`.

## Custom expanders

An expander receives a fully classified `Expression`, the shared evaluation `Context`, and returns one replacement string:

```go
custom := shellexpand.ExpanderFunc{
    ExpansionKind: shellexpand.KindDollarVariable,
    Func: func(ctx *shellexpand.Context, expression shellexpand.Expression) (string, error) {
        value, err := ctx.Value(expression.Name)
        if err != nil {
            return "", err
        }
        return strings.ToUpper(value), nil
    },
}

composer := shellexpand.MustNew(custom)
```

Use `Context.ExpandWord` to recursively process a modifier's `word`, `pattern`, or `replacement` using the same composer, environment, and options. Built-in default, alternate, assignment, and required operators expand their `word` lazily, only when their Bash condition requires it.

## Options and errors

- `WithStrictUnset()` errors on a plain reference to an unset variable. The default/alternate/assignment/required operators retain their own unset and empty rules.
- `WithMaxDepth(n)` limits nested braced expressions and recursively expanded words. The default is 1000.
- `WithUnsupportedError()` errors on a balanced `${...}` form outside this package's grammar.
- `WithUnselectedError()` errors on a supported expression whose `Kind` has no registered expander.

Parse failures are reported as `*ParseError`. Required-variable failures are `*UnsetVariableError`. Invalid glob patterns return `ErrBadPattern`. Errors from an expander are wrapped with the original expression for context and remain inspectable with `errors.Is` and `errors.As`.

## Pattern dialect and compatibility boundaries

The parameter-pattern matcher supports the basic Bash glob subset:

- `*` matches any string, including an empty string.
- `?` matches one Unicode code point.
- `[abc]`, `[a-z]`, `[!abc]`, and `[^abc]` match character classes/ranges.
- Backslash quotes the following pattern character.

This is intentionally not the complete Bash pattern language. Locale-sensitive collation, POSIX character classes such as `[[:alpha:]]`, equivalence/collating classes, extended globs (`@(…)`, `!(…)`, and friends), shell options such as `nocasematch`, array parameters, and positional/special parameters are not implemented. Range ordering follows Go's Unicode/RE2 behavior, not Bash's locale-specific collation rules. An invalid pattern returns `ErrBadPattern`.

Other intentional boundaries:

- Variable names use the ASCII identifier form `[A-Za-z_][A-Za-z0-9_]*`.
- `$1`, `$?`, `$@`, command substitutions, arithmetic expansions, process substitutions, and shell functions are not evaluated; unsupported spellings are left literal unless `WithUnsupportedError` applies to a balanced `${...}` expression.
- Quotes, shell word splitting, pathname expansion, tilde expansion, brace expansion, and general quote removal are not implemented. `\$NAME` is recognized as a literal-dollar escape. `$$` is a Bash special parameter (the shell PID), which is outside the supported grammar and is preserved unchanged; this is not a full shell lexer.
- Environment values are inserted as values; the package does not repeatedly parse variable contents as source code.
- `${!prefix*}` and `${!prefix@}` are returned as one space-delimited string. Bash's distinction between their shell word-list behavior needs a word-list result type, which this string API does not expose.
- Substring offsets and `${#NAME}` count Unicode code points. Bash's behavior can depend on the active locale/character encoding.
- Pattern replacement uses the matched text supplied by the expression's pattern; Bash's optional `patsub_replacement` behavior where unquoted `&` in the replacement expands to the matched text is not implemented.
- Empty search patterns are treated as no-ops. Zero-width-only global replacement behavior is intentionally not modeled; non-empty glob matches use longest-match, left-to-right, non-overlapping replacement.

These boundaries are deliberate: the package handles parameter-value expressions, not the full shell expansion pipeline.

## Architecture

The package uses a small set of patterns where they solve an actual problem:

- **Strategy**: each `Expander` is an interchangeable implementation for one exact `Kind`.
- **Registry/composition**: `Composer` indexes expanders by `Kind`, rather than trying handlers in a chain. This prevents an overly broad handler from partially consuming a different syntax form.
- **Adapter**: `LookupFunc` adapts `os.LookupEnv`-style functions to `Env`.
- **Functional options**: per-call behavior is configured with `WithStrictUnset`, `WithMaxDepth`, and the error options.
- **Factory functions**: named constructors create small, single-purpose built-in expanders.
- **Context passing**: `Context` carries the environment and recursive expansion behavior without global variables.
- **Recursive-descent-style scanning**: the parser recognizes complete nested `${...}` boundaries before classification; regular expressions are used only after parsing, for glob matching.

The project intentionally does not force unrelated GoF patterns onto a small library. `MapEnv` uses `sync.RWMutex` because it is a shared mutable value store; `Composer` itself is immutable after construction and can be shared concurrently.

## Tests

```sh
go test ./...
go test -race ./...
go vet ./...
```

The tests cover exact parser classification, selective dispatch, empty-vs-unset semantics, assignments, nested lazy words, Unicode substrings, pattern operations, malformed syntax, error typing, and concurrent use. A fuzz target checks that arbitrary inputs do not panic.

## Design research and references

This implementation was developed after reviewing the parser/evaluator/environment/testing approaches in:

- [`mfridman/interpolate`](https://github.com/mfridman/interpolate)
- [`a8m/envsubst`](https://github.com/a8m/envsubst)
- [`drone/envsubst`](https://github.com/drone/envsubst)
- [`ganbarodigital/go_shellexpand`](https://github.com/ganbarodigital/go_shellexpand)
- [BuildKit Dockerfile shell lexer](https://github.com/moby/buildkit/blob/master/frontend/dockerfile/shell/lex.go) and [Dockerfile variable replacement docs](https://github.com/moby/buildkit/blob/master/frontend/dockerfile/docs/reference.md)
- [GNU Bash parameter expansion](https://www.gnu.org/software/bash/manual/html_node/Shell-Parameter-Expansion.html) and [pattern matching](https://www.gnu.org/software/bash/manual/html_node/Pattern-Matching.html)
- [Go Patterns](https://go-patterns.dev/) and [Refactoring Guru's Go design-pattern examples](https://github.com/RefactoringGuru/design-patterns-go)

The cited libraries informed design and compatibility research; this project has its own parser, data model, dispatch model, and implementations.

## Publishing

The intended GitHub location is `github.com/ep0ll/shellexpand`. After creating an empty repository at that location, publish this project with:

```sh
git init -b main
git add .
git commit -m "Initial implementation"
git remote add origin https://github.com/ep0ll/shellexpand.git
git push -u origin main
```
