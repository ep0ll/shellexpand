package shellexpand

import (
	"fmt"
	"sort"
	"strings"
)

func prefixNamesExpander(kind Kind) Expander {
	return builtin(kind, func(ctx *Context, e Expression) (string, error) {
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
