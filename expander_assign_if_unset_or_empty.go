package shellexpand

// AssignIfUnsetOrEmptyExpander implements ${name:=word}.
func AssignIfUnsetOrEmptyExpander() Expander {
	return builtin(KindAssignIfUnsetOrEmpty, func(ctx *Context, e Expression) (string, error) {
		value, _ := ctx.Lookup(e.Name)
		if value != "" {
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
