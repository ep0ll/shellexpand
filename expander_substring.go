package shellexpand

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// ErrInvalidSubstringRange is returned when a negative substring length
// would make the computed end precede an in-range starting offset.
var ErrInvalidSubstringRange = errors.New("shellexpand: substring expression < 0")

// SubstringExpander implements ${name:offset} and ${name:offset:length}.
// Negative offsets count from the end; a negative length stops that many
// characters before the end. Indexes count Unicode code points, not bytes.
// Invalid UTF-8 bytes are preserved and each counts as one character.
func SubstringExpander() Expander {
	return builtin(KindSubstring, func(ctx *Context, e Expression) (string, error) {
		value, err := ctx.Value(e.Name)
		if err != nil {
			return "", err
		}
		size := utf8.RuneCountInString(value)
		start := e.Offset
		if start < 0 {
			start += size
		}
		if start < 0 || start > size {
			return "", nil
		}

		end := size
		if e.HasLength {
			switch {
			case e.Length >= 0:
				if e.Length < size-start {
					end = start + e.Length
				}
			default:
				end = size + e.Length
				if end < start {
					return "", fmt.Errorf("%w: length %d from offset %d", ErrInvalidSubstringRange, e.Length, e.Offset)
				}
			}
		}

		offsets := runeOffsets(value)
		return value[offsets[start]:offsets[end]], nil
	})
}

// runeOffsets returns byte boundaries for each decoded rune. DecodeRune
// treats an invalid UTF-8 byte as a one-byte RuneError, so slicing by these
// offsets preserves the original bytes rather than normalizing malformed UTF-8.
func runeOffsets(value string) []int {
	offsets := make([]int, 0, utf8.RuneCountInString(value)+1)
	for offset := 0; offset < len(value); {
		offsets = append(offsets, offset)
		_, width := utf8.DecodeRuneInString(value[offset:])
		offset += width
	}
	return append(offsets, len(value))
}
