# Architecture

## Goals

The implementation separates **recognition**, **classification**, and
**evaluation**. The core must not assume that expressions begin with `$`,
use braces, reference variables, or represent shell syntax. Bash is one
built-in dialect, not a dependency of the generic engine.

The project uses small Go interfaces and explicit composition. It avoids
reflection-based plugin discovery, global mutable registries, and ordered
handler chains.

## Package layout

```text
.
├── api.go, expand.go, options.go, errors.go, environment.go
├── bash.go                 # public Bash dialect constructors and kinds
├── internal/
│   ├── engine/              # syntax-neutral composer, context, contracts
│   ├── syntax/bash/         # Bash scanner and typed Bash expression data
│   ├── expanders/bash/      # Bash expansion strategies grouped by behavior
│   ├── pattern/             # shell-style glob matching
│   └── environment/         # MapEnv and ProcessEnv adapters
├── examples/customsyntax/   # non-dollar syntax extension example
└── docs/                    # architecture and implementation research
```

The root package is the supported public API. Internal packages are free to
change without creating a compatibility promise for downstream consumers.

## Core contracts

### Kind and Expression

`engine.Kind` is an open-ended string type. Applications may declare their own
kinds. The reserved `KindUnknown` is used when a syntax recognizes something
but intentionally does not classify it for evaluation.

An `Expression` stores:
- `Kind`: the exact operation to dispatch;
- `Raw`: the complete source token, used for literal preservation and errors;
- `Data`: an opaque payload owned by the recognizing syntax.

The engine has no variable name, operator, word, pattern, or replacement
fields. Such values belong in a dialect-specific payload such as
`bash.ExpressionData`, carried through `Expression.Data`.

### Syntax

A `Syntax` recognizes a complete token at a byte offset and returns a
`SyntaxMatch`. A match gives an exclusive end offset and may contain either
an expression or a literal replacement. The latter allows syntaxes to
recognize escape sequences without teaching the core what an escape means.

Multiple recognizers are composed in an ordered slice. The first successful
recognizer at an offset takes precedence. A recognizer must not consume input
when it returns `false`, and every successful match must advance the input.

### Expander

An `Expander` declares exactly one `Kind`. At composer construction, each
expander is inserted into a map keyed by kind. Empty and duplicate kinds are
rejected. Evaluation looks up exactly one expander; it never asks several
expanders whether they can handle a token.

This is stronger than a chain-of-responsibility design for overlapping
grammars: a plain variable strategy cannot claim a default-value expression
merely because both contain a variable name.

### Composer and Context

A `Composer` owns immutable syntax and expander registries and can be shared
between goroutines when its registered implementations are safe for
concurrent use. Per-call options are copied into a fresh evaluation config.

A `Context` is created for one expansion operation and carries the injected
environment, recursion depth, and composer. It provides lookup, assignment,
enumeration, and recursive word expansion without relying on package globals.
Custom expanders can use `ExpandWord` when a nested value should be evaluated
with the same active syntax and options.

## Built-in Bash dialect

The Bash parser lives in `internal/syntax/bash`. It scans nested braces and
classifies complete expressions before any expander runs. Its typed payload
stores Bash-specific names, words, patterns, replacements, and substring
bounds.

Built-in Bash strategies live in `internal/expanders/bash`, grouped into
variable access, conditional/default operations, pattern transforms, case
transforms, replacements, indirect references, and dialect registration.
Shared glob matching is separate in `internal/pattern`.

The generic engine knows only a Kind and opaque payload. It does not import
the Bash parser or Bash expander package. The public facade wires these
components together in `bash.go`.

## Environment boundary

The engine depends only on the small `Env` interface. Mutation and name
enumeration are optional capabilities exposed through `MutableEnv` and
`EnumerableEnv`. This keeps plain lookup usable with read-only sources.

`MapEnv` copies its initial values and protects mutation with an RWMutex.
`ProcessEnv` explicitly adapts the process environment using the Go standard
library. Neither expander implementations nor the generic composer directly
read process-global environment state.

## Extension checklist

When adding a dialect:
1. Define stable, non-empty Kind constants in a separate syntax package.
2. Implement a recognizer that returns complete tokens and opaque parsed data.
3. Implement expanders that claim exact kinds.
4. Add a runnable example and tests for malformed, overlapping, and nested tokens.
5. Register syntax precedence explicitly with `NewWithSyntax`.
6. Document concurrency, escaping, and unsupported syntax behavior.

See [implementation-notes.md](research/implementation-notes.md) for source
research and compatibility rationale.
