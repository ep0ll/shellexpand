package shellexpand

// DefaultIfUnsetExpander implements ${name-word}.
func DefaultIfUnsetExpander() Expander {
	return builtin(KindDefaultIfUnset, func(ctx *Context, e Expression) (string, error) {
		value, set := ctx.Lookup(e.Name)
		if set {
			return value, nil
		}
		return ctx.ExpandWord(e.Word)
	})
}
