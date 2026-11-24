package bash

// DefaultIfUnsetExpander implements ${name-word}.
func DefaultIfUnsetExpander() Expander {
	return builtin(KindDefaultIfUnset, func(ctx *Context, e ExpressionData) (string, error) {
		value, set := ctx.Lookup(e.Name)
		if set {
			return value, nil
		}
		return ctx.ExpandWord(e.Word)
	})
}

// DefaultIfUnsetOrEmptyExpander implements ${name:-word}.
func DefaultIfUnsetOrEmptyExpander() Expander {
	return builtin(KindDefaultIfUnsetOrEmpty, func(ctx *Context, e ExpressionData) (string, error) {
		value, _ := ctx.Lookup(e.Name)
		if value != "" {
			return value, nil
		}
		return ctx.ExpandWord(e.Word)
	})
}

// AssignIfUnsetExpander implements ${name=word}.
func AssignIfUnsetExpander() Expander {
	return builtin(KindAssignIfUnset, func(ctx *Context, e ExpressionData) (string, error) {
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

// AssignIfUnsetOrEmptyExpander implements ${name:=word}.
func AssignIfUnsetOrEmptyExpander() Expander {
	return builtin(KindAssignIfUnsetOrEmpty, func(ctx *Context, e ExpressionData) (string, error) {
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

// AlternateIfSetExpander implements ${name+word}.
func AlternateIfSetExpander() Expander {
	return builtin(KindAlternateIfSet, func(ctx *Context, e ExpressionData) (string, error) {
		_, set := ctx.Lookup(e.Name)
		if !set {
			return "", nil
		}
		return ctx.ExpandWord(e.Word)
	})
}

// AlternateIfSetAndNonEmptyExpander implements ${name:+word}.
func AlternateIfSetAndNonEmptyExpander() Expander {
	return builtin(KindAlternateIfSetAndNonEmpty, func(ctx *Context, e ExpressionData) (string, error) {
		value, _ := ctx.Lookup(e.Name)
		if value == "" {
			return "", nil
		}
		return ctx.ExpandWord(e.Word)
	})
}

// ErrorIfUnsetExpander implements ${name?word}.
func ErrorIfUnsetExpander() Expander {
	return builtin(KindErrorIfUnset, func(ctx *Context, e ExpressionData) (string, error) {
		value, set := ctx.Lookup(e.Name)
		if set {
			return value, nil
		}
		return "", requiredError(ctx, e, "not set")
	})
}

// ErrorIfUnsetOrEmptyExpander implements ${name:?word}.
func ErrorIfUnsetOrEmptyExpander() Expander {
	return builtin(KindErrorIfUnsetOrEmpty, func(ctx *Context, e ExpressionData) (string, error) {
		value, _ := ctx.Lookup(e.Name)
		if value != "" {
			return value, nil
		}
		return "", requiredError(ctx, e, "null or unset")
	})
}

func requiredError(ctx *Context, e ExpressionData, defaultMessage string) error {
	message, err := ctx.ExpandWord(e.Word)
	if err != nil {
		return err
	}
	if message == "" {
		message = defaultMessage
	}
	return &UnsetVariableError{Name: e.Name, Message: message}
}
