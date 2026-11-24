package shellexpand

// BracedVariableExpander expands only the plain ${name} form. Expressions
// with modifiers are parsed into different Kinds and are never matched here.
func BracedVariableExpander() Expander {
	return builtin(KindBracedVariable, func(ctx *Context, expression Expression) (string, error) {
		return ctx.Value(expression.Name)
	})
}
