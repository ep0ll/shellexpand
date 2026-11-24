package engine

import (
	"errors"
	"fmt"
)

var (
	// ErrNilExpanderFunc is returned when an ExpanderFunc has no callback.
	ErrNilExpanderFunc = errors.New("shellexpand: nil expander function")
	// ErrUnsupportedExpression is returned for a recognized but unsupported expression.
	ErrUnsupportedExpression = errors.New("unsupported expression")
	// ErrReadOnlyEnvironment is returned when an operation requires mutation.
	ErrReadOnlyEnvironment = errors.New("environment is read-only")
	// ErrVariableNameUnavailable is returned when enumeration is required but unavailable.
	ErrVariableNameUnavailable = errors.New("environment does not support variable-name enumeration")
)

// ParseError reports malformed input and its byte offset.
type ParseError struct {
	Offset     int
	Expression string
	Reason     string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("shellexpand: %s at byte %d (%q)", e.Reason, e.Offset, e.Expression)
}

// UnsetVariableError describes a required variable that is unset or empty.
type UnsetVariableError struct {
	Name    string
	Message string
}

func (e *UnsetVariableError) Error() string {
	if e == nil {
		return "shellexpand: unset variable"
	}
	return fmt.Sprintf("$%s: %s", e.Name, e.Message)
}

// UnselectedExpanderError identifies a recognized kind with no registered expander.
type UnselectedExpanderError struct {
	Kind Kind
	Raw  string
}

func (e *UnselectedExpanderError) Error() string {
	return fmt.Sprintf("no expander registered for %s (%s)", e.Kind, e.Raw)
}

// UnsupportedExpressionError identifies a balanced but unsupported expression.
type UnsupportedExpressionError struct{ Raw string }

func (e *UnsupportedExpressionError) Error() string {
	return fmt.Sprintf("%s %s", ErrUnsupportedExpression, e.Raw)
}

// Unwrap supports errors.Is(err, ErrUnsupportedExpression).
func (e *UnsupportedExpressionError) Unwrap() error { return ErrUnsupportedExpression }
