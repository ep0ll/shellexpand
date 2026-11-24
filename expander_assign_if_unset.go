package shellexpand

// AssignIfUnsetExpander implements ${name=word}.
func AssignIfUnsetExpander() Expander {
	return builtin(KindAssignIfUnset, func(ctx *Context, e Expression) (string, error) {
		value, set := ctx.Lookup(e.Name)
		if set {
			return value, nil
		}
		value, err := ctx.ExpandWord(e.Word)
		if err != nil {
			return "", err
		}
		if err := ctx.Assign(e.Name, value); err != nil {
			return "", err
		}
		return value, nil
	})
}
