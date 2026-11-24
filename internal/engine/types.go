// Package engine contains the syntax-independent expansion engine.
package engine

// Kind identifies an expansion operation. Kinds are open-ended: applications
// may define their own values for any syntax they register.
type Kind string

// KindUnknown is reserved for syntaxes that recognize an expression but
// cannot classify its operation.
const KindUnknown Kind = "unknown"

// Expression is one complete expression recognized by a Syntax. Data carries
// syntax-specific parsed state and is intentionally opaque to the engine.
type Expression struct {
	Kind Kind
	Raw  string
	Data any
}

// Expander evaluates one syntactic kind. Implementations should be safe for
// concurrent use; request-specific state belongs in Context or Expression.
type Expander interface {
	Kind() Kind
	Expand(ctx *Context, expression Expression) (string, error)
}

// ExpanderFunc adapts a function to Expander.
type ExpanderFunc struct {
	ExpansionKind Kind
	Func          func(*Context, Expression) (string, error)
}

// Kind returns the kind handled by f.
func (f ExpanderFunc) Kind() Kind { return f.ExpansionKind }

// Expand evaluates expression using f.Func.
func (f ExpanderFunc) Expand(ctx *Context, expression Expression) (string, error) {
	if f.Func == nil {
		return "", ErrNilExpanderFunc
	}
	return f.Func(ctx, expression)
}
