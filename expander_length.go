package shellexpand

import (
	"strconv"
	"unicode/utf8"
)

// LengthExpander implements ${#name}. Length is counted in Unicode code
// points, matching Bash in a UTF-8 locale rather than Go's byte length.
func LengthExpander() Expander {
	return builtin(KindLength, func(ctx *Context, e Expression) (string, error) {
		value, err := ctx.Value(e.Name)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(utf8.RuneCountInString(value)), nil
	})
}
