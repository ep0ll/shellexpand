package shellexpand

// IndirectVariableExpander implements ${!name}: the value of name is treated
// as the name of a second variable.
func IndirectVariableExpander() Expander {
	return builtin(KindIndirectVariable, func(ctx *Context, e Expression) (string, error) {
		target, err := ctx.Value(e.Name)
		if err != nil {
			return "", err
		}
		if target == "" {
			return "", nil
		}
		return ctx.Value(target)
	})
}
