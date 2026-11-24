package shellexpand

// ErrorIfUnsetOrEmptyExpander implements ${name:?word}.
func ErrorIfUnsetOrEmptyExpander() Expander {
	return builtin(KindErrorIfUnsetOrEmpty, func(ctx *Context, e Expression) (string, error) {
		value, _ := ctx.Lookup(e.Name)
		if value != "" {
			return value, nil
		}
		return "", requiredError(ctx, e, "null or unset")
	})
}
