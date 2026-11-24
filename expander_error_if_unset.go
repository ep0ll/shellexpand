package shellexpand

// ErrorIfUnsetExpander implements ${name?word}.
func ErrorIfUnsetExpander() Expander {
	return builtin(KindErrorIfUnset, func(ctx *Context, e Expression) (string, error) {
		value, set := ctx.Lookup(e.Name)
		if set {
			return value, nil
		}
		return "", requiredError(ctx, e, "not set")
	})
}
