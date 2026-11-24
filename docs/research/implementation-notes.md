# Implementation research notes

These notes record the design research behind the library. They summarize
source-level patterns and important semantic distinctions; they are not a
claim of complete Bash compatibility.

## Bash parameter expansion

The Bash manual distinguishes unset values from empty values. The colon
forms `:-`, `:=`, `:?`, and `:+` treat either state as absent; the
non-colon forms only treat an unset variable as absent. These cases must be
parsed as different operations, not inferred by an expander after a broad
variable match.

Prefix and suffix removal support shortest and longest matches
(`#`/`##`, `%`/`%%`). Pattern substitution has first/global and
prefix/suffix-anchored forms. Case modification may use an optional pattern.
Indirect variable expansion and prefix-name enumeration also have distinct
semantics. Conditional words should only be expanded when the corresponding
branch is selected.

The supported library subset is intentionally narrower than the full shell
grammar. Locale-sensitive collation, POSIX character classes, extended
globs, arrays, positional parameters, shell options, quote rules, and
word-list semantics are outside the string API's scope.

## Go implementation research

The design reviewed several open-source implementations:

- [mfridman/interpolate](https://github.com/mfridman/interpolate): recursive
  descent parsing, expression/evaluator separation, typed AST nodes,
  set-versus-empty lookup, lazy word evaluation, depth limits, corpus tests,
  and fuzz tests.
- [a8m/envsubst](https://github.com/a8m/envsubst): separation of lexing and
  parsing, explicit node types, and lexer state transitions.
- [drone/envsubst](https://github.com/drone/envsubst): parser/evaluator
  separation, function-style extensions, and matching helpers.
- [ganbarodigital/go_shellexpand](https://github.com/ganbarodigital/go_shellexpand):
  environment callbacks, variable parsing, operation-specific callbacks,
  and separation of expansion phases.
- [BuildKit's Dockerfile shell lexer](https://github.com/moby/buildkit/blob/master/frontend/dockerfile/shell/lex.go):
  stateful scanning for escapes, quotes, and nested shell-like constructs.
  Dockerfile variable replacement is a distinct phase from command shell
  execution.

## Architectural decisions

The library uses a syntax-neutral `Composer`, an ordered list of `Syntax`
recognizers, and a registry from exact `Kind` to `Expander`. This makes
custom syntaxes possible without adding grammar branches to the core. It
also avoids accidental partial matches that can arise when expansion
handlers are tried sequentially.

`Expression.Data` is deliberately opaque. The Bash dialect owns its typed
payload, while a custom template language may store an entirely different
AST node or token metadata in the same field.

Small Go interfaces express optional environment capabilities. Functional
options carry per-call behavior. Factories construct independent strategies.
The design applies these patterns only where they simplify substitution,
testing, or ownership; it does not attempt to force a pattern onto every type.

## OSS project conventions

The repository's directory layout and maintenance files take inspiration
from established Go projects including Kubernetes, Juju, Dapr, and
CoreOS/go-systemd. Only conventions useful for a small reusable library are
adopted: clear package boundaries, contributor and security policies,
ownership metadata, CI, CodeQL, dependency automation, formatting/lint
checks, and opt-in local Git hooks.
