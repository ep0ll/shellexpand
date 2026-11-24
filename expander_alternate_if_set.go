package shellexpand

// AlternateIfSetExpander implements ${name+word}.
func AlternateIfSetExpander() Expander {
	return builtin(KindAlternateIfSet, func(ctx *Context, e Expression) (string, error) {
		_, set := ctx.Lookup(e.Name)
		if !set {
			return "", nil
		}
		return ctx.ExpandWord(e.Word)
	})
}
