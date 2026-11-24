package bash

import (
	"regexp"
	"strings"
	"unicode/utf8"

	glob "github.com/ep0ll/shellexpand/internal/pattern"
)

type replaceMode int

const (
	replaceFirstMode replaceMode = iota
	replaceAllMode
	replacePrefixMode
	replaceSuffixMode
)

func replaceExpander(kind Kind, mode replaceMode) Expander {
	return builtin(kind, func(ctx *Context, e ExpressionData) (string, error) {
		value, err := ctx.Value(e.Name)
		if err != nil {
			return "", err
		}
		pattern, err := ctx.ExpandWord(e.Pattern)
		if err != nil {
			return "", err
		}
		replacement, err := ctx.ExpandWord(e.Replacement)
		if err != nil {
			return "", err
		}
		if pattern == "" {
			return value, nil
		}

		if mode == replacePrefixMode {
			matcher, err := glob.CompilePrefix(pattern, true)
			if err != nil {
				return "", err
			}
			indices := matcher.FindStringIndex(value)
			if indices == nil || indices[0] != 0 || indices[1] == 0 && value != "" {
				return value, nil
			}
			return replacement + value[indices[1]:], nil
		}
		if mode == replaceSuffixMode {
			if _, err := glob.Compile(pattern); err != nil {
				return "", err
			}
			reversedPattern := glob.ReverseGlobPattern(pattern)
			matcher, err := glob.CompilePrefix(reversedPattern, true)
			if err != nil {
				return "", err
			}
			reversedValue := glob.ReverseRunes(value)
			indices := matcher.FindStringIndex(reversedValue)
			if indices == nil || indices[0] != 0 || indices[1] == 0 && value != "" {
				return value, nil
			}
			return value[:len(value)-indices[1]] + replacement, nil
		}

		matcher, err := glob.CompileSearch(pattern)
		if err != nil {
			return "", err
		}
		if value == "" {
			if matcher.MatchString("") {
				return replacement, nil
			}
			return value, nil
		}
		return replaceMatches(matcher, value, replacement, mode == replaceAllMode), nil
	})
}

// replaceMatches replaces leftmost, longest, non-overlapping non-empty
// matches. Zero-width matches are skipped to avoid insertion loops.
func replaceMatches(matcher *regexp.Regexp, value, replacement string, all bool) string {
	firstStart, firstEnd, ok := findNonEmptyMatch(matcher, value, 0)
	if !ok {
		return value
	}

	var out strings.Builder
	out.Grow(len(value))
	last := 0
	start, end := firstStart, firstEnd
	for {
		out.WriteString(value[last:start])
		out.WriteString(replacement)
		last = end
		if !all {
			break
		}
		start, end, ok = findNonEmptyMatch(matcher, value, last)
		if !ok {
			break
		}
	}
	out.WriteString(value[last:])
	return out.String()
}

func findNonEmptyMatch(matcher *regexp.Regexp, value string, offset int) (start, end int, ok bool) {
	for offset <= len(value) {
		indices := matcher.FindStringIndex(value[offset:])
		if indices == nil {
			return 0, 0, false
		}
		start = offset + indices[0]
		end = offset + indices[1]
		if end > start {
			return start, end, true
		}
		if start >= len(value) {
			return 0, 0, false
		}
		_, width := utf8.DecodeRuneInString(value[start:])
		offset = start + width
	}
	return 0, 0, false
}

// ReplaceFirstExpander implements ${name/pattern/replacement}.
func ReplaceFirstExpander() Expander {
	return replaceExpander(KindReplaceFirst, replaceFirstMode)
}

// ReplaceAllExpander implements ${name//pattern/replacement}.
func ReplaceAllExpander() Expander {
	return replaceExpander(KindReplaceAll, replaceAllMode)
}

// ReplacePrefixExpander implements ${name/#pattern/replacement}.
func ReplacePrefixExpander() Expander {
	return replaceExpander(KindReplacePrefix, replacePrefixMode)
}

// ReplaceSuffixExpander implements ${name/%pattern/replacement}.
func ReplaceSuffixExpander() Expander {
	return replaceExpander(KindReplaceSuffix, replaceSuffixMode)
}
