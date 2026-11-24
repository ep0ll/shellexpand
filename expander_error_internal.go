package shellexpand

func requiredError(ctx *Context, e Expression, defaultMessage string) error {
	message, err := ctx.ExpandWord(e.Word)
	if err != nil {
		return err
	}
	if message == "" {
		message = defaultMessage
	}
	return &UnsetVariableError{Name: e.Name, Message: message}
}
