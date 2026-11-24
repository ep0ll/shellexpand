package bash

import "fmt"

type conditionalMode uint8

const (
	conditionalDefault conditionalMode = iota
	conditionalAssign
	conditionalAlternate
	conditionalError
)

// conditionalExpander implements the shared unset/empty decision used by
// Bash's default, assignment, alternate, and required-value operators.
func conditionalExpander(kind Kind, mode conditionalMode, emptyIsUnset bool) Expander {
	return builtin(kind, func(ctx *Context, e ExpressionData) (string, error) {
		value, set := ctx.Lookup(e.Name)
		unset := !set || (emptyIsUnset && value == "")

		switch mode {
		case conditionalDefault:
			if !unset {
				return value, nil
			}
			return ctx.ExpandWord(e.Word)

		case conditionalAssign:
			if !unset {
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

		case conditionalAlternate:
			if unset {
				return "", nil
			}
			return ctx.ExpandWord(e.Word)

		case conditionalError:
			if !unset {
				return value, nil
			}
			message := "not set"
			if emptyIsUnset {
				message = "null or unset"
			}
			return "", requiredError(ctx, e, message)

		default:
			return "", fmt.Errorf("shellexpand: invalid conditional expander mode %d", mode)
		}
	})
}

// DefaultIfUnsetExpander implements ${name-word}.
func DefaultIfUnsetExpander() Expander {
	return conditionalExpander(KindDefaultIfUnset, conditionalDefault, false)
}

// DefaultIfUnsetOrEmptyExpander implements ${name:-word}.
func DefaultIfUnsetOrEmptyExpander() Expander {
	return conditionalExpander(KindDefaultIfUnsetOrEmpty, conditionalDefault, true)
}

// AssignIfUnsetExpander implements ${name=word}.
func AssignIfUnsetExpander() Expander {
	return conditionalExpander(KindAssignIfUnset, conditionalAssign, false)
}

// AssignIfUnsetOrEmptyExpander implements ${name:=word}.
func AssignIfUnsetOrEmptyExpander() Expander {
	return conditionalExpander(KindAssignIfUnsetOrEmpty, conditionalAssign, true)
}

// AlternateIfSetExpander implements ${name+word}.
func AlternateIfSetExpander() Expander {
	return conditionalExpander(KindAlternateIfSet, conditionalAlternate, false)
}

// AlternateIfSetAndNonEmptyExpander implements ${name:+word}.
func AlternateIfSetAndNonEmptyExpander() Expander {
	return conditionalExpander(KindAlternateIfSetAndNonEmpty, conditionalAlternate, true)
}

// ErrorIfUnsetExpander implements ${name?word}.
func ErrorIfUnsetExpander() Expander {
	return conditionalExpander(KindErrorIfUnset, conditionalError, false)
}

// ErrorIfUnsetOrEmptyExpander implements ${name:?word}.
func ErrorIfUnsetOrEmptyExpander() Expander {
	return conditionalExpander(KindErrorIfUnsetOrEmpty, conditionalError, true)
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
