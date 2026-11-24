package shellexpand

import (
	"fmt"
	"strings"
	"testing"
)

type customToken struct{ name string }

func TestCustomSyntaxAndExpanderUseNonDollarDelimiters(t *testing.T) {
	t.Parallel()

	const customKind Kind = "percent_delimited_variable"
	syntax := SyntaxFunc(func(input string, offset int, _ ParseOptions) (SyntaxMatch, bool, error) {
		if offset+4 > len(input) || input[offset:offset+2] != "%%" {
			return SyntaxMatch{}, false, nil
		}
		relativeEnd := strings.Index(input[offset+2:], "%%")
		if relativeEnd < 0 {
			return SyntaxMatch{}, false, fmt.Errorf("unterminated custom variable at byte %d", offset)
		}
		end := offset + 2 + relativeEnd + 2
		name := input[offset+2 : end-2]
		expression := Expression{Kind: customKind, Raw: input[offset:end], Data: customToken{name: name}}
		return SyntaxMatch{End: end, Expression: &expression}, true, nil
	})
	expander := ExpanderFunc{
		ExpansionKind: customKind,
		Func: func(ctx *Context, expression Expression) (string, error) {
			token, ok := expression.Data.(customToken)
			if !ok {
				return "", fmt.Errorf("unexpected syntax payload %T", expression.Data)
			}
			return ctx.Value(token.name)
		},
	}
	composer := MustNewWithSyntax([]Syntax{syntax}, expander)
	got, err := composer.Expand("image=%%IMAGE%% tag=%%TAG%%", NewMapEnv(map[string]string{
		"IMAGE": "alpine",
		"TAG":   "3.22",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if want := "image=alpine tag=3.22"; got != want {
		t.Fatalf("Expand() = %q, want %q", got, want)
	}
}

func TestCustomKindsAreOpenAndSyntaxesComposeByPrecedence(t *testing.T) {
	t.Parallel()

	first := SyntaxFunc(func(input string, offset int, _ ParseOptions) (SyntaxMatch, bool, error) {
		if strings.HasPrefix(input[offset:], "[[") {
			if end := strings.Index(input[offset+2:], "]]"); end >= 0 {
				return SyntaxMatch{End: offset + end + 4, Expression: &Expression{Kind: "first", Data: input[offset+2 : offset+end+2]}}, true, nil
			}
		}
		return SyntaxMatch{}, false, nil
	})
	second := SyntaxFunc(func(input string, offset int, _ ParseOptions) (SyntaxMatch, bool, error) {
		if strings.HasPrefix(input[offset:], "<<") {
			if end := strings.Index(input[offset+2:], ">>"); end >= 0 {
				return SyntaxMatch{End: offset + end + 4, Expression: &Expression{Kind: "second", Data: input[offset+2 : offset+end+2]}}, true, nil
			}
		}
		return SyntaxMatch{}, false, nil
	})
	expanders := []Expander{
		ExpanderFunc{ExpansionKind: "first", Func: func(_ *Context, e Expression) (string, error) { return "A:" + e.Data.(string), nil }},
		ExpanderFunc{ExpansionKind: "second", Func: func(_ *Context, e Expression) (string, error) { return "B:" + e.Data.(string), nil }},
	}
	composer := MustNewWithSyntax([]Syntax{first, second}, expanders...)
	got, err := composer.Expand("[[one]] <<two>>", nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := "A:one B:two"; got != want {
		t.Fatalf("Expand() = %q, want %q", got, want)
	}
}

func TestSyntaxCanReplaceMatchedTextWithEmptyString(t *testing.T) {
	t.Parallel()

	syntax := SyntaxFunc(func(input string, offset int, _ ParseOptions) (SyntaxMatch, bool, error) {
		if strings.HasPrefix(input[offset:], "<remove>") {
			return SyntaxMatch{End: offset + len("<remove>"), Replace: true}, true, nil
		}
		return SyntaxMatch{}, false, nil
	})
	composer := MustNewWithSyntax([]Syntax{syntax})
	got, err := composer.Expand("before<remove>after", nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := "beforeafter"; got != want {
		t.Fatalf("Expand() = %q, want %q", got, want)
	}
}

func TestCustomSyntaxCanComposeWithBash(t *testing.T) {
	t.Parallel()

	const customKind Kind = "example.bracket_variable"
	customSyntax := SyntaxFunc(func(input string, offset int, _ ParseOptions) (SyntaxMatch, bool, error) {
		if !strings.HasPrefix(input[offset:], "[[") {
			return SyntaxMatch{}, false, nil
		}
		end := strings.Index(input[offset+2:], "]]")
		if end < 0 {
			return SyntaxMatch{}, false, fmt.Errorf("unterminated bracket expression")
		}
		end += offset + 4
		expression := Expression{Kind: customKind, Data: input[offset+2 : end-2]}
		return SyntaxMatch{End: end, Expression: &expression}, true, nil
	})
	customExpander := ExpanderFunc{
		ExpansionKind: customKind,
		Func: func(ctx *Context, expression Expression) (string, error) {
			name, ok := expression.Data.(string)
			if !ok {
				return "", fmt.Errorf("unexpected payload %T", expression.Data)
			}
			return ctx.Value(name)
		},
	}
	composer := MustNewWithSyntax([]Syntax{BashSyntax(), customSyntax}, append(DefaultExpanders(), customExpander)...)
	got, err := composer.Expand("$A/${B}/[[C]]", NewMapEnv(map[string]string{"A": "one", "B": "two", "C": "three"}))
	if err != nil {
		t.Fatal(err)
	}
	if want := "one/two/three"; got != want {
		t.Fatalf("Expand() = %q, want %q", got, want)
	}
}

func TestOverlappingSyntaxesRespectRegistrationOrder(t *testing.T) {
	t.Parallel()

	makeSyntax := func(kind Kind) Syntax {
		return SyntaxFunc(func(input string, offset int, _ ParseOptions) (SyntaxMatch, bool, error) {
			if !strings.HasPrefix(input[offset:], "[[") {
				return SyntaxMatch{}, false, nil
			}
			relativeEnd := strings.Index(input[offset+2:], "]]")
			if relativeEnd < 0 {
				return SyntaxMatch{}, false, fmt.Errorf("unterminated token")
			}
			end := offset + 2 + relativeEnd + 2
			expression := Expression{Kind: kind, Data: input[offset+2 : end-2]}
			return SyntaxMatch{End: end, Expression: &expression}, true, nil
		})
	}
	first := makeSyntax("priority.first")
	second := makeSyntax("priority.second")
	expanders := []Expander{
		ExpanderFunc{ExpansionKind: "priority.first", Func: func(_ *Context, e Expression) (string, error) {
			return "first:" + e.Data.(string), nil
		}},
		ExpanderFunc{ExpansionKind: "priority.second", Func: func(_ *Context, e Expression) (string, error) {
			return "second:" + e.Data.(string), nil
		}},
	}

	tests := []struct {
		name     string
		syntaxes []Syntax
		want     string
	}{
		{name: "first has precedence", syntaxes: []Syntax{first, second}, want: "first:value"},
		{name: "second has precedence when registered first", syntaxes: []Syntax{second, first}, want: "second:value"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			composer := MustNewWithSyntax(test.syntaxes, expanders...)
			got, err := composer.Expand("[[value]]", nil)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("Expand() = %q, want %q", got, test.want)
			}
		})
	}
}
