package shellexpand

import "github.com/ep0ll/shellexpand/internal/engine"

// WithStrictUnset returns an option that errors on lookups of unset variables.
func WithStrictUnset() Option { return engine.WithStrictUnset() }

// WithMaxDepth sets the maximum recursive expansion depth.
func WithMaxDepth(depth int) Option { return engine.WithMaxDepth(depth) }

// WithUnsupportedError errors when a syntax recognizes an unknown expression kind.
func WithUnsupportedError() Option { return engine.WithUnsupportedError() }

// WithUnselectedError errors when no expander is registered for a recognized kind.
func WithUnselectedError() Option { return engine.WithUnselectedError() }
