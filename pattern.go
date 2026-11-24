package shellexpand

import "github.com/ep0ll/shellexpand/internal/pattern"

// MatchGlob reports whether value matches the built-in shell-style glob syntax.
func MatchGlob(expression, value string) (bool, error) {
	return pattern.Match(expression, value)
}
