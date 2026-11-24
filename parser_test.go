package shellexpand

import (
	"errors"
	"testing"
)

func TestParserClassifiesEveryExpressionPrecisely(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		kind  Kind
		name  string
	}{
		{"$NAME", KindDollarVariable, "NAME"},
		{"${NAME}", KindBracedVariable, "NAME"},
		{"${NAME:-word}", KindDefaultIfUnsetOrEmpty, "NAME"},
		{"${NAME-word}", KindDefaultIfUnset, "NAME"},
		{"${NAME:=word}", KindAssignIfUnsetOrEmpty, "NAME"},
		{"${NAME=word}", KindAssignIfUnset, "NAME"},
		{"${NAME:?word}", KindErrorIfUnsetOrEmpty, "NAME"},
		{"${NAME?word}", KindErrorIfUnset, "NAME"},
		{"${NAME:+word}", KindAlternateIfSetAndNonEmpty, "NAME"},
		{"${NAME+word}", KindAlternateIfSet, "NAME"},
		{"${#NAME}", KindLength, "NAME"},
		{"${NAME:1:2}", KindSubstring, "NAME"},
		{"${NAME#*.}", KindRemoveShortestPrefix, "NAME"},
		{"${NAME##*.}", KindRemoveLongestPrefix, "NAME"},
		{"${NAME%.*}", KindRemoveShortestSuffix, "NAME"},
		{"${NAME%%.*}", KindRemoveLongestSuffix, "NAME"},
		{"${NAME/a/b}", KindReplaceFirst, "NAME"},
		{"${NAME//a/b}", KindReplaceAll, "NAME"},
		{"${NAME/#a/b}", KindReplacePrefix, "NAME"},
		{"${NAME/%a/b}", KindReplaceSuffix, "NAME"},
		{"${NAME^a}", KindUppercaseFirst, "NAME"},
		{"${NAME^^a}", KindUppercaseAll, "NAME"},
		{"${NAME,a}", KindLowercaseFirst, "NAME"},
		{"${NAME,,a}", KindLowercaseAll, "NAME"},
		{"${!NAME}", KindIndirectVariable, "NAME"},
		{"${!PRE_*}", KindIndirectPrefixNamesStar, "PRE_"},
		{"${!PRE_@}", KindIndirectPrefixNamesAt, "PRE_"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			parts, err := parse(test.input, 100)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if len(parts) != 1 || parts[0].expression == nil {
				t.Fatalf("expected one expression, got %#v", parts)
			}
			got := *parts[0].expression
			if got.Kind != test.kind || got.Name != test.name {
				t.Errorf("got kind/name %q/%q, want %q/%q", got.Kind, got.Name, test.kind, test.name)
			}
		})
	}
}

func TestParserPreservesUnsupportedAndLiteralSyntax(t *testing.T) {
	t.Parallel()

	input := `$1 $(echo hi) $((1 + 2)) ${#} ${name@Q} $ ${bad` // final brace is intentionally unclosed
	if _, err := parse(input, 100); err == nil {
		t.Fatal("expected unclosed braced expansion error")
	}

	input = `$1 $(echo hi) $((1 + 2)) ${#} ${name@Q} $`
	got, err := Expand(input, nil)
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	if got != input {
		t.Fatalf("got %q, want unchanged %q", got, input)
	}
}

func TestParserEscapesAndNestedBraces(t *testing.T) {
	t.Parallel()

	env := NewMapEnv(map[string]string{"SET": "value", "EMPTY": ""})
	tests := []struct{ input, want string }{
		{`$$SET`, `$$SET`},
		{`\$SET`, `$SET`},
		{`${EMPTY:-${SET}}`, `value`},
		{`${EMPTY:-literal\}brace}`, `literal}brace`},
		{`${EMPTY:-${EMPTY2:-literal\}brace}}`, `literal}brace`},
		{`${EMPTY:-$${SET\}x}`, `$${SET}x`},
		{`${EMPTY:-\$SET}`, `$SET`},
		{`${EMPTY:-$${SET}}`, `$${SET}`},
	}
	for _, test := range tests {
		got, err := Expand(test.input, env)
		if err != nil {
			t.Errorf("Expand(%q): %v", test.input, err)
			continue
		}
		if got != test.want {
			t.Errorf("Expand(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestParserSubstringSyntaxErrors(t *testing.T) {
	t.Parallel()
	for _, input := range []string{`${NAME:}`, `${NAME:x}`, `${NAME:1:x}`, `${NAME:1:2:3}`} {
		t.Run(input, func(t *testing.T) {
			if _, err := parse(input, 100); err == nil {
				t.Fatal("expected parse error")
			}
		})
	}
}

func TestParserNestingLimit(t *testing.T) {
	t.Parallel()
	input := "${X:-" + "${X:-" + "${X:-value}}}"
	if _, err := parse(input, 2); err == nil {
		t.Fatal("expected depth limit error")
	}
	if _, err := parse("${X:-${X:-value}}", 2); err != nil {
		t.Fatalf("unexpected error at exact depth: %v", err)
	}
}

func TestParseErrorIsTyped(t *testing.T) {
	t.Parallel()
	_, err := parse("${A", 10)
	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("error %v is not *ParseError", err)
	}
}

func TestComposerRejectsTypedNilExpander(t *testing.T) {
	t.Parallel()
	var expander *builtinExpander
	if _, err := New(expander); err == nil {
		t.Fatal("expected typed nil expander to be rejected")
	}
}
