package shellexpand

import (
	"strings"
	"unicode/utf8"
)

func caseExpander(kind Kind, upper, first bool) Expander {
	return builtin(kind, func(ctx *Context, e Expression) (string, error) {
		value, err := ctx.Value(e.Name)
		if err != nil {
			return "", err
		}
		pattern, err := ctx.ExpandWord(e.Pattern)
		if err != nil {
			return "", err
		}

		var matcher interface{ MatchString(string) bool }
		if pattern != "" {
			matcher, err = compileGlob(pattern)
			if err != nil {
				return "", err
			}
		}

		var out strings.Builder
		out.Grow(len(value))
		firstMatchPending := first
		for offset := 0; offset < len(value); {
			r, width := utf8.DecodeRuneInString(value[offset:])
			segment := value[offset : offset+width]
			invalidUTF8 := r == utf8.RuneError && width == 1 && value[offset] >= utf8.RuneSelf
			matched := pattern == ""
			if !matched {
				matched = matcher.MatchString(segment)
			}
			if matched && (!first || firstMatchPending) {
				if invalidUTF8 {
					out.WriteString(segment)
				} else if upper {
					out.WriteString(strings.ToUpper(segment))
				} else {
					out.WriteString(strings.ToLower(segment))
				}
				if first {
					firstMatchPending = false
				}
			} else {
				out.WriteString(segment)
			}
			offset += width
		}
		return out.String(), nil
	})
}
