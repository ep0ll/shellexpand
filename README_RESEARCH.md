# Research notes and implementation rationale

This document records the source-level study behind the implementation. It is intentionally a compatibility study, not a claim that this library parses or behaves like a complete Bash interpreter.

## GNU Bash parameter-expansion semantics

The reference manual establishes several distinctions that must survive parsing:

- `-`, `:-`, `=`, `:=`, `?`, `:?`, `+`, and `:+` are different operators. The colon forms additionally treat an empty value like an unset value; the non-colon forms do not.
- `#` / `##` remove the shortest / longest matching prefix; `%` / `%%` do the same for suffixes.
- Pattern substitution chooses the longest match, replacing only the first match unless the operator contains `//`. `/#` and `/%` anchor matching at the beginning and end.
- Case modification may take an optional pattern. With no pattern, Bash's default is a single-character wildcard.
- Indirect expansion and prefix-name enumeration are distinct from ordinary variable lookup. In particular, `${!prefix*}` / `${!prefix@}` return names, rather than the value of a variable literally named `prefix`.
- Substitution words are themselves expanded. Default and alternate words are conditional; they should not be evaluated if their branch is not selected.

The implementation keeps set-vs-empty state in `Env.Lookup`, expands conditional words lazily, and models each visible syntax form as a separate `Kind`. GNU Bash also has locale-dependent pattern behavior, shell options, arrays, quoting rules, and list-valued expansions which this string library intentionally does not claim to implement.

## Existing Go packages

### `mfridman/interpolate`

Important source files reviewed:

- `parser.go`: recursive-descent scanner/parser for nested `${...}` expressions. It creates a small expression tree and imposes a nesting-depth limit rather than trying to parse nested expressions with a single regular expression.
- `expression.go`: an evaluator interface implemented by distinct value/default/required/alternate/length/substring expansion nodes. Evaluation is separate from syntax parsing; required-word evaluation and conditional defaults happen lazily.
- `env.go`: `Env.Lookup(name) (value, ok)` differentiates an unset variable from a variable explicitly set to the empty string. `LookupFunc` adapts `os.LookupEnv`; map/list environment constructors copy and normalize inputs.
- `interpolate.go`: public `Expand`, `Variables`, options, and typed unset errors. Options are applied per call.
- `doc.go`: documents supported shell syntax and concurrency expectations.

Tests reviewed include `interpolate_test.go`, `parser_test.go`, `env_test.go`, `env_internal_test.go`, `corpus_test.go`, `fuzz_test.go`, and `concurrency_test.go`. The Bash-generated corpus is particularly useful because it stores fixtures generated from Bash and does not need Bash at test time. Fuzzing checks parser/evaluator invariants, while concurrent-use tests establish safe use when the provided environment is safe. The package's deliberate omissions (notably pattern removal, assignment, replacement, and case changes) make it a useful parser/evaluator baseline, but not a full implementation of the user's listed forms.

### `a8m/envsubst`

Important source files reviewed:

- `envsubst.go`: public string/byte/file helpers and restriction-based entry points.
- `parse/lex.go`: state-machine lexer. It tracks substitution depth, handles `$` / `${` tokens, recognizes the default operators, and can preserve digit-leading references literally.
- `parse/parse.go`: parser that turns lexer output into nodes and supports a quick or collect-errors mode.
- `parse/node.go`: `TextNode`, `VariableNode`, and `SubstitutionNode`, each with evaluation behavior.
- `cmd/envsubst/main.go`: CLI wiring for strict/unset/empty/digit policies and input/output.

Tests reviewed include `envsubst_test.go`, `parse/lex_test.go`, and `parse/parse_test.go`. They exercise escaped dollars, nested/default forms, unset/empty policy, and parser failure modes. The useful design separation is lexer -> parser -> node evaluation. The implementation's restrictions are useful, but the parser/evaluator covers a smaller operator set than required here. The lexer uses a goroutine and channel to emit tokens; this repository uses a synchronous scanner to avoid an extra goroutine/channel lifecycle for a single-pass string parse.

### `drone/envsubst`

Important source files reviewed:

- `template.go`: parsed `Template` plus a per-execution state containing writer, current node, and mapping callback.
- `eval.go`: convenience entry points using a caller mapping or process environment.
- `funcs.go`: operation-specific transformations and function-name dispatch.
- `parse/parse.go`: recursive parser for template text, nested function arguments, separators, and escapes.
- `parse/node.go`: `TextNode`, `FuncNode`, and `ListNode` model.
- `parse/scan.go`: rune-aware scanner with mode bits, accept functions, token boundaries, and escape rules.
- `path/match.go`: Go `path.Match`-style wildcard matcher.

Tests reviewed include `eval_test.go`, `funcs_test.go`, `parse/parse_test.go`, and the path matcher source. Its AST execution approach is useful, and the table-driven tests cover many requested transformations. There are also semantic hazards worth avoiding: a generic default function can become a silent fallback for unrecognized operations; the `path.Match` dialect treats slash specially, unlike Bash parameter-pattern matching where slash is not a special separator; and byte-based substring/length operations can diverge from Bash character counting. This repository uses strict `Kind` lookup, explicitly named expanders, a Bash-inspired glob subset that matches slash, and rune/code-point operations.

### `ganbarodigital/go_shellexpand`

Important source files reviewed:

- `ExpansionCallbacks.go`: injected callbacks for assigning and looking up variables, home directories, and variable-name enumeration.
- `expand.go`: staged shell pipeline: brace expansion, tilde expansion, parameter/variable expansion, then arithmetic/quote-removal stages (some stages are separate/limited in the package).
- `expandParams.go`: large parameter-expansion dispatcher and operation implementations. It resolves indirection, accesses values through callbacks, and then applies an operator-specific handler.
- `matchVar.go`: finds the end of `$name` / `${...}` with nested-brace and escape awareness.
- `matchParam.go` plus the specialized name/positional/special matchers: classifies shell parameter names.
- `matchParamOp.go`: maps operator spelling to internal operation identifiers.
- `parseParams.go`: converts the complete raw parameter expression into a descriptor with operation kind, parts, and indirection flag.
- `expandParams_test.go`, `matchVar_test.go`, `matchParamOp_test.go`, and `parseParams_test.go`: tests include nested brace recognition, escaped braces, recognized operators, malformed names, positional/special parameters, and unsupported operation handling.

The callback interface is a valuable dependency-inversion pattern, but the broad shell feature set makes the dispatch and parameter descriptor more complex. The new library retains the injected environment idea but separates lookup, mutation, and enumeration into small interfaces. It does not implement brace/tilde/arithmetic/quote-removal because those are distinct shell expansion phases and outside this parameter-expansion API.

## BuildKit / Dockerfile

`frontend/dockerfile/shell/lex.go` was reviewed as an implementation reference. BuildKit's `Lex` handles shell-like words, quotes, escapes, variable substitution, matched/unmatched variable tracking, and optionally post-substitution word splitting. It is deliberately not a complete Bash parser; the code has a distinct `processDollar` path for the supported `${...}` operators. The API is stateful and a `Lex` instance is not concurrently reusable.

The Dockerfile reference also distinguishes builder-side variable replacement in specific Dockerfile instructions from command-shell expansion in shell-form `RUN`, `CMD`, and `ENTRYPOINT`. Those are separate layers. This project implements neither Dockerfile instruction substitution nor shell word tokenization; it implements a reusable parameter-expansion layer only.

## Go design-pattern research and chosen patterns

The Go Patterns catalog includes practical Go-specific approaches such as functional options, configuration-struct factories, configurable objects, lazy initialization, and map-plus-`sync.RWMutex` caching. Refactoring Guru's Go catalog covers classic patterns including Strategy, Adapter, Composite, Decorator, and Chain of Responsibility.

Patterns applied here because they directly solve the selection/composition problem:

- **Strategy**: every syntax form has its own `Expander` implementation/factory and can be selected independently.
- **Registry / composition**: `Composer` indexes handlers by exact `Kind`. A chain-of-responsibility dispatcher was deliberately rejected: a broad `${NAME}` handler in a first-match chain could consume a prefix of `${NAME-default}`. Parsing the entire expression and indexing by one exact kind makes that class of bug structurally impossible.
- **Adapter**: `LookupFunc` adapts `os.LookupEnv`-style functions to `Env` without binding expanders to the process environment.
- **Functional options**: strict-unset behavior, depth limits, and unsupported/unselected-expression policies are call-scoped.
- **Factories**: named constructors expose independent built-in strategies without forcing users to manipulate internal function values.
- **Context passing**: `Context` supplies lookup, optional assignment/enumeration, and recursive word expansion explicitly.
- **Synchronization by ownership**: `MapEnv` owns a copied map and protects mutation with `sync.RWMutex`; `Composer` is immutable after construction. No global mutable registry/cache is required.

The implementation intentionally avoids forcing Decorator, Visitor, Singleton, or other patterns into code where they would add indirection without solving a real requirement.

## Implementation structure

Each concrete expansion form has its own source file and public factory. Shared mechanics live in focused internal files (glob compilation, replacement mechanics, and related helpers). The parser produces a complete `Expression` before the composer dispatches it. An unselected kind is either preserved as its original raw spelling or returned as a typed error by option; it is never sent to a different expander as a fallback.

## Compatibility boundary

This is a Bash-inspired subset, not a shell. It deliberately omits command/arithmetic/process substitution, positional and special parameters, array semantics, shell quoting and word splitting, tilde/brace expansion, pathname expansion, locale-sensitive character classes/collation, POSIX bracket sub-classes, extended globbing, and advanced replacement-string behavior (including optional `patsub_replacement`). Pattern matching uses `*`, `?`, simple bracket classes/ranges, negation, and backslash escapes. Unknown balanced `${...}` forms remain unchanged unless `WithUnsupportedError` is requested.


## Standard-library reuse

The implementation intentionally uses Go's standard library for the generic mechanics: `os.LookupEnv` / `os.Setenv` / `os.Environ` through `ProcessEnv`, `strings.Builder` and `strings.Cut` for string assembly and environment-list parsing, `strconv` and `unicode/utf8` for numeric and Unicode handling, `regexp` for matching the translated Bash-glob subset, `sync.RWMutex` for `MapEnv`, and `slices` / `sort` for deterministic name ordering. A custom scanner/parser remains necessary because Go's `os.Expand` only handles basic variable references and cannot distinguish and parse the nested Bash parameter-operator grammar. The standard `path.Match` / `filepath.Match` APIs are not substituted for Bash parameter patterns: their path-oriented `*` semantics do not match across `/`, while Bash parameter-pattern operators are not path matching.
