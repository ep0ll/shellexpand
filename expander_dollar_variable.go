package shellexpand

// DollarVariableExpander expands only the unbraced $name form.
func DollarVariableExpander() Expander {
	return builtin(KindDollarVariable, func(ctx *Context, expression Expression) (string, error) {
		return ctx.Value(expression.Name)
	})
}
