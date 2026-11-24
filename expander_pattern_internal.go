package shellexpand

func trimExpander(kind Kind, prefix, longest bool) Expander {
	return builtin(kind, func(ctx *Context, e Expression) (string, error) {
		value, err := ctx.Value(e.Name)
		if err != nil {
			return "", err
		}
		pattern, err := ctx.ExpandWord(e.Pattern)
		if err != nil {
			return "", err
		}
		if _, err := compileGlob(pattern); err != nil {
			return "", err
		}

		matchValue := value
		matchPattern := pattern
		if !prefix {
			matchValue = reverseRunes(value)
			matchPattern = reverseGlobPattern(pattern)
		}
		matcher, err := compileGlobPrefix(matchPattern, longest)
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
