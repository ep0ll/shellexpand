// Package shellexpand provides composable, syntax-specific expanders for a
// subset of Bash parameter expansion.
//
// Use Expand for the complete built-in set, or construct a Composer with only
// the expansion forms your application permits. For example:
//
//	composer := shellexpand.MustNew(
//		shellexpand.DollarVariableExpander(),
//		shellexpand.BracedVariableExpander(),
//	)
//	result, err := composer.Expand("$NAME ${NAME}", env)
//
// Pass [ProcessEnv]{} to use os.LookupEnv, os.Setenv, and os.Environ, or
// use [LookupFunc] to adapt a lookup function such as os.LookupEnv. Prefer
// [MapEnv] when assignments should not mutate the process environment.
//
// The parser classifies an entire ${...} expression before selecting an
// expander. Therefore ${NAME:-fallback} is not partially consumed by the
// BracedVariableExpander when default expansion is not enabled; the original
// expression is preserved unchanged (or reported with WithUnselectedError).
//
// This package implements parameter-value transformations only. It does not
// execute command substitutions, arithmetic expressions, shell functions,
// pathname expansion, word splitting, arrays, positional parameters, or
// general shell quoting rules. Pattern operators support the basic Bash glob
// metacharacters *, ?, and bracket character classes.
package shellexpand
