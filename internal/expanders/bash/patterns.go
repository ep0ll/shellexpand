package bash

import glob "github.com/ep0ll/shellexpand/internal/pattern"

func trimExpander(kind Kind, prefix, longest bool) Expander {
	return builtin(kind, func(ctx *Context, e ExpressionData) (string, error) {
		value, err := ctx.Value(e.Name)
		if err != nil {
			return "", err
		}
		pattern, err := ctx.ExpandWord(e.Pattern)
		if err != nil {
			return "", err
		}
		// Suffix matching reverses the pattern. Validate it before token
		// reversal; prefix matching is validated by CompilePrefix below.
		if !prefix {
			if _, err := glob.Compile(pattern); err != nil {
				return "", err
			}
		}

		matchValue := value
		matchPattern := pattern
		if !prefix {
			matchValue = glob.ReverseRunes(value)
			matchPattern = glob.ReverseGlobPattern(pattern)
		}
		matcher, err := glob.CompilePrefix(matchPattern, longest)
		if err != nil {
			return "", err
		}
		indices := matcher.FindStringIndex(matchValue)
		if indices == nil || indices[0] != 0 || indices[1] == 0 {
			return value, nil
		}
		matchedBytes := indices[1]
		if prefix {
			return value[matchedBytes:], nil
		}
		return value[:len(value)-matchedBytes], nil
	})
}

// RemoveShortestPrefixExpander implements ${name#pattern}.
func RemoveShortestPrefixExpander() Expander {
	return trimExpander(KindRemoveShortestPrefix, true, false)
}

// RemoveLongestPrefixExpander implements ${name##pattern}.
func RemoveLongestPrefixExpander() Expander {
	return trimExpander(KindRemoveLongestPrefix, true, true)
}

// RemoveShortestSuffixExpander implements ${name%pattern}.
func RemoveShortestSuffixExpander() Expander {
	return trimExpander(KindRemoveShortestSuffix, false, false)
}

// RemoveLongestSuffixExpander implements ${name%%pattern}.
func RemoveLongestSuffixExpander() Expander {
	return trimExpander(KindRemoveLongestSuffix, false, true)
}
