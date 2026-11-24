package shellexpand

import (
	"github.com/ep0ll/shellexpand/internal/engine"
	bashexpand "github.com/ep0ll/shellexpand/internal/expanders/bash"
	"github.com/ep0ll/shellexpand/internal/pattern"
)

var (
	// ErrNilExpanderFunc is returned by an ExpanderFunc without a callback.
	ErrNilExpanderFunc = engine.ErrNilExpanderFunc
	// ErrUnsupportedExpression identifies an expression recognized as unsupported.
	ErrUnsupportedExpression = engine.ErrUnsupportedExpression
	// ErrReadOnlyEnvironment indicates that an expansion requires mutation of a read-only Env.
	ErrReadOnlyEnvironment = engine.ErrReadOnlyEnvironment
	// ErrVariableNameUnavailable indicates that Env does not support name enumeration.
	ErrVariableNameUnavailable = engine.ErrVariableNameUnavailable
	// ErrBadPattern indicates that a shell-style pattern is malformed.
	ErrBadPattern = pattern.ErrBadPattern
	// ErrInvalidSubstringRange identifies an invalid negative-length substring range.
	ErrInvalidSubstringRange = bashexpand.ErrInvalidSubstringRange
)
