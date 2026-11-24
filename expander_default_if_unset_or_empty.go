package shellexpand

// DefaultIfUnsetOrEmptyExpander implements ${name:-word}.
func DefaultIfUnsetOrEmptyExpander() Expander {
	return builtin(KindDefaultIfUnsetOrEmpty, func(ctx *Context, e Expression) (string, error) {
		value, _ := ctx.Lookup(e.Name)
		if value != "" {
			return value, nil
		}
		return ctx.ExpandWord(e.Word)
	})
}
