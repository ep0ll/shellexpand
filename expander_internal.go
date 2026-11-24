package shellexpand

type builtinExpander struct {
	kind Kind
	fn   func(*Context, Expression) (string, error)
}

func (e builtinExpander) Kind() Kind { return e.kind }
func (e builtinExpander) Expand(ctx *Context, expression Expression) (string, error) {
	return e.fn(ctx, expression)
}

func builtin(kind Kind, fn func(*Context, Expression) (string, error)) Expander {
	return builtinExpander{kind: kind, fn: fn}
}
