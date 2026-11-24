package bash

import (
	"fmt"
	"sort"
	"strings"
)

func prefixNamesExpander(kind Kind) Expander {
	return builtin(kind, func(ctx *Context, e ExpressionData) (string, error) {
		names, ok := ctx.Names()
		if !ok {
			return "", fmt.Errorf("shellexpand: %w", ErrVariableNameUnavailable)
		}
		matched := make([]string, 0)
		for _, name := range names {
			if strings.HasPrefix(name, e.Name) {
				matched = append(matched, name)
			}
		}
		sort.Strings(matched)
		return strings.Join(matched, " "), nil
	})
}

// IndirectVariableExpander implements ${!name}: the value of name is treated
// as the name of a second variable.
func IndirectVariableExpander() Expander {
	return builtin(KindIndirectVariable, func(ctx *Context, e ExpressionData) (string, error) {
		target, err := ctx.Value(e.Name)
		if err != nil {
			return "", err
		}
		if target == "" {
			return "", nil
		}
		return ctx.Value(target)
	})
}

// IndirectPrefixNamesStarExpander implements ${!prefix*}.
func IndirectPrefixNamesStarExpander() Expander {
	return prefixNamesExpander(KindIndirectPrefixNamesStar)
}

// IndirectPrefixNamesAtExpander implements ${!prefix@}. Like the star form,
// this joins names with spaces because this package returns strings rather
// than Bash's shell-word list.
func IndirectPrefixNamesAtExpander() Expander {
	return prefixNamesExpander(KindIndirectPrefixNamesAt)
}
