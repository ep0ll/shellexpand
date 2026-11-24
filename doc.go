// Package shellexpand provides a composable variable expansion engine.
//
// The core engine is independent of any shell syntax. Applications register
// one or more Syntax recognizers and Expander strategies keyed by arbitrary
// Kind values. The built-in Bash dialect is available through BashSyntax,
// DefaultExpanders, and Expand.
//
// Expansion is syntax-directed: a recognizer returns a complete expression
// before its Kind is dispatched. This prevents a broad handler from consuming
// part of an expression belonging to another operator or dialect.
package shellexpand
