package shellexpand

// AlternateIfSetAndNonEmptyExpander implements ${name:+word}.
func AlternateIfSetAndNonEmptyExpander() Expander {
	return builtin(KindAlternateIfSetAndNonEmpty, func(ctx *Context, e Expression) (string, error) {
		value, _ := ctx.Lookup(e.Name)
		if value == "" {
			return "", nil
		}
		return ctx.ExpandWord(e.Word)
	})
}
