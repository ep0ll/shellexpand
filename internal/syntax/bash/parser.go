package bash

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ep0ll/shellexpand/internal/engine"
)

// Syntax recognizes Bash-style dollar parameter expressions. Its zero value is ready to use.
type Syntax struct{}

// NewSyntax returns the Bash parameter-expansion syntax recognizer.
func NewSyntax() Syntax { return Syntax{} }

// Match implements engine.Syntax. It recognizes one complete token at offset.
func (Syntax) Match(input string, offset int, options engine.ParseOptions) (engine.SyntaxMatch, bool, error) {
	if offset < 0 || offset >= len(input) {
		return engine.SyntaxMatch{}, false, nil
	}
	if input[offset] == '\\' && offset+1 < len(input) && input[offset+1] == '$' {
		return engine.SyntaxMatch{End: offset + 2, Text: "$", Replace: true}, true, nil
	}
	if input[offset] != '$' {
		return engine.SyntaxMatch{}, false, nil
	}
	if offset+1 < len(input) && input[offset+1] == '$' {
		return engine.SyntaxMatch{End: offset + 2, Text: "$$"}, true, nil
	}
	if offset+1 < len(input) && input[offset+1] == '{' {
		maxDepth := options.MaxDepth
		if maxDepth <= 0 {
			maxDepth = 1000
		}
		end, err := findBraceEnd(input, offset, maxDepth)
		if err != nil {
			return engine.SyntaxMatch{}, false, err
		}
		raw := input[offset:end]
		body := input[offset+2 : end-1]
		if body == "" {
			return engine.SyntaxMatch{}, false, &engine.ParseError{Offset: offset, Expression: raw, Reason: "empty parameter expansion"}
		}
		expression, err := parseBraced(raw, body, offset)
		if err != nil {
			return engine.SyntaxMatch{}, false, err
		}
		return engine.SyntaxMatch{End: end, Expression: &expression}, true, nil
	}
	if offset+1 < len(input) && isNameStart(input[offset+1]) {
		end := offset + 2
		for end < len(input) && isNameContinue(input[end]) {
			end++
		}
		expression := Expression{Kind: KindDollarVariable, Raw: input[offset:end], Data: ExpressionData{Name: input[offset+1 : end]}}
		return engine.SyntaxMatch{End: end, Expression: &expression}, true, nil
	}
	return engine.SyntaxMatch{}, false, nil
}

func findBraceEnd(input string, start, maxDepth int) (int, error) {
	depth := 1
	for i := start + 2; i < len(input); i++ {
		// Escaped closing braces do not end this expression. Escaped dollars
		// and doubled dollars do not begin nested expressions either.
		if input[i] == '\\' && i+1 < len(input) {
			i++
			continue
		}
		if input[i] == '$' && i+1 < len(input) && input[i+1] == '$' {
			i++
			continue
		}
		if input[i] == '$' && i+1 < len(input) && input[i+1] == '{' {
			depth++
			if depth > maxDepth {
				return 0, fmt.Errorf("shellexpand: parameter expansions nested deeper than %d levels", maxDepth)
			}
			i++
			continue
		}
		if input[i] == '}' {
			depth--
			if depth == 0 {
				return i + 1, nil
			}
		}
	}
	return 0, &engine.ParseError{Offset: start, Expression: input[start:], Reason: "missing closing brace"}
}

func parseBraced(raw, body string, offset int) (Expression, error) {
	unknown := func() Expression { return Expression{Kind: KindUnknown, Raw: raw} }

	// Length uses a leading #, while # and ## after a name remove prefixes.
	if strings.HasPrefix(body, "#") {
		name := body[1:]
		if validName(name) {
			return Expression{Kind: KindLength, Raw: raw, Data: ExpressionData{Name: name}}, nil
		}
		return unknown(), nil
	}

	// Indirection has its own grammar and must not fall through to ${name}.
	if strings.HasPrefix(body, "!") {
		indirect := body[1:]
		if strings.HasSuffix(indirect, "*") && validPrefix(indirect[:len(indirect)-1]) {
			return Expression{Kind: KindIndirectPrefixNamesStar, Raw: raw, Data: ExpressionData{Name: indirect[:len(indirect)-1]}}, nil
		}
		if strings.HasSuffix(indirect, "@") && validPrefix(indirect[:len(indirect)-1]) {
			return Expression{Kind: KindIndirectPrefixNamesAt, Raw: raw, Data: ExpressionData{Name: indirect[:len(indirect)-1]}}, nil
		}
		if validName(indirect) {
			return Expression{Kind: KindIndirectVariable, Raw: raw, Data: ExpressionData{Name: indirect}}, nil
		}
		return unknown(), nil
	}

	name, rest := takeName(body)
	if name == "" {
		return unknown(), nil
	}
	if rest == "" {
		return Expression{Kind: KindBracedVariable, Raw: raw, Data: ExpressionData{Name: name}}, nil
	}

	// Test two-character operators before one-character operators. The
	// syntactic classification happens before dispatch; no simple-variable
	// expander can ever claim these expressions.
	for _, operator := range []struct {
		text string
		kind Kind
	}{
		{":-", KindDefaultIfUnsetOrEmpty},
		{":=", KindAssignIfUnsetOrEmpty},
		{":?", KindErrorIfUnsetOrEmpty},
		{":+", KindAlternateIfSetAndNonEmpty},
		{"-", KindDefaultIfUnset},
		{"=", KindAssignIfUnset},
		{"?", KindErrorIfUnset},
		{"+", KindAlternateIfSet},
	} {
		if strings.HasPrefix(rest, operator.text) {
			return Expression{Kind: operator.kind, Raw: raw, Data: ExpressionData{Name: name, Word: unescapeBracedWord(rest[len(operator.text):])}}, nil
		}
	}

	switch rest[0] {
	case ':':
		return parseSubstring(raw, name, rest[1:], offset)
	case '#':
		if strings.HasPrefix(rest, "##") {
			return Expression{Kind: KindRemoveLongestPrefix, Raw: raw, Data: ExpressionData{Name: name, Pattern: unescapeBracedWord(rest[2:])}}, nil
		}
		return Expression{Kind: KindRemoveShortestPrefix, Raw: raw, Data: ExpressionData{Name: name, Pattern: unescapeBracedWord(rest[1:])}}, nil
	case '%':
		if strings.HasPrefix(rest, "%%") {
			return Expression{Kind: KindRemoveLongestSuffix, Raw: raw, Data: ExpressionData{Name: name, Pattern: unescapeBracedWord(rest[2:])}}, nil
		}
		return Expression{Kind: KindRemoveShortestSuffix, Raw: raw, Data: ExpressionData{Name: name, Pattern: unescapeBracedWord(rest[1:])}}, nil
	case '/':
		return parseReplacement(raw, name, rest)
	case '^':
		if strings.HasPrefix(rest, "^^") {
			return Expression{Kind: KindUppercaseAll, Raw: raw, Data: ExpressionData{Name: name, Pattern: unescapeBracedWord(rest[2:])}}, nil
		}
		return Expression{Kind: KindUppercaseFirst, Raw: raw, Data: ExpressionData{Name: name, Pattern: unescapeBracedWord(rest[1:])}}, nil
	case ',':
		if strings.HasPrefix(rest, ",,") {
			return Expression{Kind: KindLowercaseAll, Raw: raw, Data: ExpressionData{Name: name, Pattern: unescapeBracedWord(rest[2:])}}, nil
		}
		return Expression{Kind: KindLowercaseFirst, Raw: raw, Data: ExpressionData{Name: name, Pattern: unescapeBracedWord(rest[1:])}}, nil
	default:
		return unknown(), nil
	}
}

func parseSubstring(raw, name, rest string, offset int) (Expression, error) {
	offsetText, lengthText, hasLength := strings.Cut(rest, ":")
	if !hasLength && strings.TrimSpace(offsetText) == "" {
		return Expression{}, &engine.ParseError{Offset: offset, Expression: raw, Reason: "substring offset is required"}
	}
	parsedOffset, err := parseSubstringInteger(offsetText)
	if err != nil {
		return Expression{}, &engine.ParseError{Offset: offset, Expression: raw, Reason: "invalid substring offset: " + err.Error()}
	}
	data := ExpressionData{Name: name, Offset: parsedOffset, HasLength: hasLength}
	if hasLength {
		data.Length, err = parseSubstringInteger(lengthText)
		if err != nil {
			return Expression{}, &engine.ParseError{Offset: offset, Expression: raw, Reason: "invalid substring length: " + err.Error()}
		}
	}
	return Expression{Kind: KindSubstring, Raw: raw, Data: data}, nil
}

func parseSubstringInteger(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	// Bash accepts the parenthesized literal form used to disambiguate a
	// negative substring offset from the :- default-value operator.
	if strings.HasPrefix(value, "(") && strings.HasSuffix(value, ")") {
		value = strings.TrimSpace(value[1 : len(value)-1])
	}
	return strconv.Atoi(value)
}

func parseReplacement(raw, name, rest string) (Expression, error) {
	kind := KindReplaceFirst
	rest = rest[1:]
	if strings.HasPrefix(rest, "/") {
		kind = KindReplaceAll
		rest = rest[1:]
	} else if strings.HasPrefix(rest, "#") {
		kind = KindReplacePrefix
		rest = rest[1:]
	} else if strings.HasPrefix(rest, "%") {
		kind = KindReplaceSuffix
		rest = rest[1:]
	}

	pattern, replacement, hasReplacement := splitUnescapedSlash(rest)
	if !hasReplacement {
		replacement = ""
	}
	return Expression{
		Kind: kind,
		Raw:  raw,
		Data: ExpressionData{
			Name:        name,
			Pattern:     unescapeBracedWord(pattern),
			Replacement: unescapeReplacementWord(replacement),
		},
	}, nil
}

// splitUnescapedSlash splits on the first unescaped slash. The backslash is
// retained so the pattern matcher can interpret \/ as a literal slash.
func splitUnescapedSlash(value string) (before, after string, found bool) {
	depth := 0
	for i := 0; i < len(value); i++ {
		if value[i] == '\\' {
			i++
			continue
		}
		if value[i] == '$' && i+1 < len(value) && value[i+1] == '$' {
			i++
			continue
		}
		if value[i] == '$' && i+1 < len(value) && value[i+1] == '{' {
			depth++
			i++
			continue
		}
		if value[i] == '}' && depth > 0 {
			depth--
			continue
		}
		if value[i] == '/' && depth == 0 {
			return value[:i], value[i+1:], true
		}
	}
	return value, "", false
}

// unescapeBracedWord removes the backslash used by Bash to include a literal
// closing brace in a ${...} word. Other backslashes remain available to the
// nested expansion parser and glob matcher.
func unescapeBracedWord(value string) string {
	if !strings.Contains(value, `\}`) {
		return value
	}
	var out strings.Builder
	out.Grow(len(value))
	depth := 0
	for i := 0; i < len(value); {
		if value[i] == '\\' && i+1 < len(value) {
			switch value[i+1] {
			case '\\', '$':
				out.WriteString(value[i : i+2])
				i += 2
				continue
			case '}':
				if depth == 0 {
					out.WriteByte('}')
					i += 2
					continue
				}
				out.WriteString(`\}`)
				i += 2
				continue
			}
		}
		if value[i] == '$' && i+1 < len(value) && value[i+1] == '$' {
			out.WriteString("$$")
			i += 2
			continue
		}
		if value[i] == '$' && i+1 < len(value) && value[i+1] == '{' {
			depth++
			out.WriteString("${")
			i += 2
			continue
		}
		if value[i] == '}' && depth > 0 {
			depth--
		}
		out.WriteByte(value[i])
		i++
	}
	return out.String()
}

// unescapeReplacementWord also removes escaped slash delimiters from the raw
// replacement word. This is applied to source text before nested parameter
// expansion, so a slash produced by an environment variable remains data.
func unescapeReplacementWord(value string) string {
	value = unescapeBracedWord(value)
	return strings.ReplaceAll(value, `\/`, `/`)
}

func takeName(value string) (name, rest string) {
	if value == "" || !isNameStart(value[0]) {
		return "", value
	}
	end := 1
	for end < len(value) && isNameContinue(value[end]) {
		end++
	}
	return value[:end], value[end:]
}

func validName(name string) bool {
	parsed, rest := takeName(name)
	return parsed != "" && rest == ""
}

func validPrefix(prefix string) bool {
	if prefix == "" {
		return true
	}
	return validName(prefix)
}

func isNameStart(b byte) bool {
	return b == '_' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

func isNameContinue(b byte) bool {
	return isNameStart(b) || b >= '0' && b <= '9'
}
