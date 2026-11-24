// Package shellexpand provides a composable, syntax-agnostic expansion engine
// and a built-in Bash parameter-expansion dialect.
package shellexpand

import "github.com/ep0ll/shellexpand/internal/engine"

type (
	// Kind identifies an expansion operation and is open to application-defined values.
	Kind = engine.Kind
	// Expression is the syntax-independent representation of one recognized token.
	Expression = engine.Expression
	// Expander evaluates expressions belonging to one exact Kind.
	Expander = engine.Expander
	// ExpanderFunc adapts a function to the Expander interface.
	ExpanderFunc = engine.ExpanderFunc
	// Syntax recognizes complete tokens at input byte offsets.
	Syntax = engine.Syntax
	// SyntaxFunc adapts a function to the Syntax interface.
	SyntaxFunc = engine.SyntaxFunc
	// SyntaxMatch describes a token recognized by a Syntax.
	SyntaxMatch = engine.SyntaxMatch
	// ParseOptions contains limits shared with syntax recognizers.
	ParseOptions = engine.ParseOptions
	// Composer combines ordered syntax recognizers with an exact-kind expander registry.
	Composer = engine.Composer
	// Context provides environment operations and controlled recursive expansion.
	Context = engine.Context
	// Option configures one expansion call.
	Option = engine.Option
	// Env provides variable lookup with an explicit set/unset distinction.
	Env = engine.Env
	// MutableEnv extends Env with variable assignment.
	MutableEnv = engine.MutableEnv
	// EnumerableEnv extends Env with variable-name enumeration.
	EnumerableEnv = engine.EnumerableEnv
	// LookupFunc adapts an os.LookupEnv-style function to Env.
	LookupFunc = engine.LookupFunc
	// ParseError describes malformed input and its byte offset.
	ParseError = engine.ParseError
	// UnsetVariableError describes a required variable that is unset or empty.
	UnsetVariableError = engine.UnsetVariableError
	// UnselectedExpanderError identifies a recognized Kind without a registered Expander.
	UnselectedExpanderError = engine.UnselectedExpanderError
	// UnsupportedExpressionError identifies a recognized but unsupported expression.
	UnsupportedExpressionError = engine.UnsupportedExpressionError
)

// New creates a composer with the built-in Bash syntax and exactly the
// expanders supplied. Unregistered expression kinds remain literal.
func New(expanders ...Expander) (*Composer, error) {
	return engine.New([]Syntax{BashSyntax()}, expanders...)
}

// NewWithSyntax creates a composer for arbitrary syntaxes and expansion kinds.
// Syntax order defines precedence when multiple recognizers match at one byte.
func NewWithSyntax(syntaxes []Syntax, expanders ...Expander) (*Composer, error) {
	return engine.New(syntaxes, expanders...)
}

// MustNew is like New but panics when the expander registry is invalid.
func MustNew(expanders ...Expander) *Composer {
	composer, err := New(expanders...)
	if err != nil {
		panic(err)
	}
	return composer
}

// MustNewWithSyntax is like NewWithSyntax but panics on invalid registrations.
func MustNewWithSyntax(syntaxes []Syntax, expanders ...Expander) *Composer {
	composer, err := NewWithSyntax(syntaxes, expanders...)
	if err != nil {
		panic(err)
	}
	return composer
}
