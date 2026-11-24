package pattern

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ErrBadPattern is returned for malformed Bash-style glob patterns used by
// parameter pattern operators.
var ErrBadPattern = errors.New("shellexpand: malformed pattern")

func Match(pattern, value string) (bool, error) {
	compiled, err := Compile(pattern)
	if err != nil {
		return false, err
	}
	return compiled.MatchString(value), nil
}

// compileGlob builds an anchored matcher for a complete string.
func Compile(pattern string) (*regexp.Regexp, error) {
	return compileExpression(pattern, true, true, true)
}

// compileGlobPrefix builds a matcher anchored at the beginning but not the
// end. A lazy wildcard selects the shortest matching prefix; a greedy one
// selects the longest matching prefix.
func CompilePrefix(pattern string, greedy bool) (*regexp.Regexp, error) {
	return compileExpression(pattern, greedy, true, false)
}

// compileGlobSearch builds an unanchored matcher. Greedy '*' quantifiers make
// the match at the earliest position as long as possible, which is the
// matching rule used by Bash's parameter replacement operators.
func CompileSearch(pattern string) (*regexp.Regexp, error) {
	return compileExpression(pattern, true, false, false)
}

func compileExpression(pattern string, greedy, anchorStart, anchorEnd bool) (*regexp.Regexp, error) {
	runes := []rune(pattern)
	var out strings.Builder
	out.WriteString("(?s)")
	if anchorStart {
		out.WriteByte('^')
	}
	star := ".*?"
	if greedy {
		star = ".*"
	}

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch r {
		case '*':
			out.WriteString(star)
		case '?':
			out.WriteByte('.')
		case '\\':
			if i+1 >= len(runes) {
				return nil, ErrBadPattern
			}
			i++
			out.WriteString(regexp.QuoteMeta(string(runes[i])))
		case '[':
			end := findClassEnd(runes, i)
			if end < 0 {
				return nil, ErrBadPattern
			}
			class := append([]rune(nil), runes[i+1:end]...)
			out.WriteByte('[')
			if len(class) > 0 && (class[0] == '!' || class[0] == '^') {
				out.WriteByte('^')
				class = class[1:]
			}
			for j := 0; j < len(class); j++ {
				r := class[j]
				if r == '\\' {
					if j+1 >= len(class) {
						return nil, ErrBadPattern
					}
					j++
					writeRegexClassLiteral(&out, class[j])
					continue
				}
				if r == ']' || r == '^' || r == '[' {
					writeRegexClassLiteral(&out, r)
					continue
				}
				out.WriteRune(r)
			}
			out.WriteByte(']')
			i = end
		default:
			out.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	if anchorEnd {
		out.WriteByte('$')
	}
	compiled, err := regexp.Compile(out.String())
	if err != nil {
		return nil, ErrBadPattern
	}
	return compiled, nil
}

// findClassEnd returns the closing bracket of a shell character class. A ]
// in the first class position (after an optional negator) is treated as a
// literal member. Backslash quotes the next rune.
func findClassEnd(pattern []rune, start int) int {
	i := start + 1
	if i < len(pattern) && (pattern[i] == '!' || pattern[i] == '^') {
		i++
	}
	if i < len(pattern) && pattern[i] == ']' {
		i++
	}
	for i < len(pattern) {
		switch pattern[i] {
		case '\\':
			i += 2
		case '[':
			// POSIX character, collating, and equivalence classes use
			// nested forms such as [[:alpha:]], [[.ch.]], and [[=a=]].
			// They are outside this package's deliberately small glob subset.
			if i+1 < len(pattern) && (pattern[i+1] == ':' || pattern[i+1] == '.' || pattern[i+1] == '=') {
				return -1
			}
			i++
		case ']':
			return i
		default:
			i++
		}
	}
	return -1
}

// writeRegexClassLiteral uses a hexadecimal rune escape so characters which
// have special meaning in RE2 character classes remain literal.
func writeRegexClassLiteral(out *strings.Builder, r rune) {
	const hex = "0123456789abcdef"
	out.WriteString(`\x{`)
	var digits [8]byte
	i := len(digits)
	for n := int(r); n > 0; n >>= 4 {
		i--
		digits[i] = hex[n&0xf]
	}
	if i == len(digits) {
		i--
		digits[i] = '0'
	}
	out.Write(digits[i:])
	out.WriteByte('}')
}

// reverseGlobPattern reverses glob tokens, not the characters inside an
// escape or bracket expression. Reversing both the pattern and input turns
// suffix matching into prefix matching without changing character classes.
// The caller validates the pattern before using this helper.
func ReverseGlobPattern(pattern string) string {
	runes := []rune(pattern)
	tokens := make([]string, 0, len(runes))
	for i := 0; i < len(runes); {
		start := i
		switch runes[i] {
		case '\\':
			i += 2
		case '[':
			end := findClassEnd(runes, i)
			i = end + 1
		default:
			i++
		}
		tokens = append(tokens, string(runes[start:i]))
	}
	var out strings.Builder
	out.Grow(len(pattern))
	for i := len(tokens) - 1; i >= 0; i-- {
		out.WriteString(tokens[i])
	}
	return out.String()
}

// reverseRunes reverses Unicode code points while retaining the original byte
// sequences, including malformed UTF-8 bytes.
func ReverseRunes(value string) string {
	offsets := RuneOffsets(value)
	var out strings.Builder
	out.Grow(len(value))
	for i := len(offsets) - 2; i >= 0; i-- {
		out.WriteString(value[offsets[i]:offsets[i+1]])
	}
	return out.String()
}

// RuneOffsets returns byte offsets for Unicode code points, preserving invalid UTF-8 bytes.
func RuneOffsets(value string) []int {
	offsets := make([]int, 0, len(value)+1)
	for offset := 0; offset < len(value); {
		offsets = append(offsets, offset)
		_, width := utf8.DecodeRuneInString(value[offset:])
		offset += width
	}
	return append(offsets, len(value))
}
