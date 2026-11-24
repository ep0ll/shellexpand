package bash

import (
	"strings"
	"unicode/utf8"

	glob "github.com/ep0ll/shellexpand/internal/pattern"
)

func caseExpander(kind Kind, upper, first bool) Expander {
	return builtin(kind, func(ctx *Context, e ExpressionData) (string, error) {
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
			matcher, err = glob.Compile(pattern)
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

// UppercaseFirstExpander implements ${name^pattern}.
func UppercaseFirstExpander() Expander {
	return caseExpander(KindUppercaseFirst, true, true)
}

// UppercaseAllExpander implements ${name^^pattern}.
func UppercaseAllExpander() Expander {
	return caseExpander(KindUppercaseAll, true, false)
}

// LowercaseFirstExpander implements ${name,pattern}.
func LowercaseFirstExpander() Expander {
	return caseExpander(KindLowercaseFirst, false, true)
}

// LowercaseAllExpander implements ${name,,pattern}.
func LowercaseAllExpander() Expander {
	return caseExpander(KindLowercaseAll, false, false)
}
