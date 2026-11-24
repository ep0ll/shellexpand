package engine

import (
	"errors"
	"strings"
	"testing"
)

func TestComposerDispatchesByExactCustomKind(t *testing.T) {
	t.Parallel()

	const kind Kind = "test.non_shell_token"
	syntax := SyntaxFunc(func(input string, offset int, _ ParseOptions) (SyntaxMatch, bool, error) {
		if !strings.HasPrefix(input[offset:], "<%") {
			return SyntaxMatch{}, false, nil
		}
		end := strings.Index(input[offset+2:], "%>")
		if end < 0 {
			return SyntaxMatch{}, false, errors.New("unterminated test token")
		}
		end += offset + 4
		return SyntaxMatch{
			End:        end,
			Expression: &Expression{Kind: kind, Data: input[offset+2 : end-2]},
		}, true, nil
	})
	expander := ExpanderFunc{
		ExpansionKind: kind,
		Func: func(_ *Context, expression Expression) (string, error) {
			value, ok := expression.Data.(string)
			if !ok {
				return "", errors.New("unexpected data payload")
			}
			if expression.Raw != "<%name%>" {
				return "", errors.New("composer did not populate original source")
			}
			return strings.ToUpper(value), nil
		},
	}
	composer, err := New([]Syntax{syntax}, expander)
	if err != nil {
		t.Fatal(err)
	}
	got, err := composer.Expand("value=<%name%>", nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := "value=NAME"; got != want {
		t.Fatalf("Expand() = %q, want %q", got, want)
	}
}

func TestComposerValidatesSyntaxRangesAndRawValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		match SyntaxMatch
		input string
	}{
		{name: "no progress", match: SyntaxMatch{End: 0}, input: "x"},
		{name: "past input", match: SyntaxMatch{End: 2}, input: "x"},
		{
			name:  "raw mismatch",
			input: "x",
			match: SyntaxMatch{End: 1, Expression: &Expression{Kind: "test", Raw: "y"}},
		},
		{
			name:  "empty kind",
			input: "x",
			match: SyntaxMatch{End: 1, Expression: &Expression{}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			syntax := SyntaxFunc(func(string, int, ParseOptions) (SyntaxMatch, bool, error) {
				return test.match, true, nil
			})
			composer := MustNew([]Syntax{syntax})
			if _, err := composer.Expand(test.input, nil); err == nil {
				t.Fatal("expected invalid syntax match error")
			}
		})
	}
}

func TestComposerReportsUnselectedCustomKind(t *testing.T) {
	t.Parallel()

	syntax := SyntaxFunc(func(input string, offset int, _ ParseOptions) (SyntaxMatch, bool, error) {
		if !strings.HasPrefix(input[offset:], "@@") {
			return SyntaxMatch{}, false, nil
		}
		return SyntaxMatch{End: offset + 2, Expression: &Expression{Kind: "custom.unregistered"}}, true, nil
	})
	composer := MustNew([]Syntax{syntax})
	if _, err := composer.Expand("@@", nil, WithUnselectedError()); err == nil {
		t.Fatal("expected unselected custom-kind error")
	}
}

func TestNewRejectsDuplicateKinds(t *testing.T) {
	t.Parallel()

	first := ExpanderFunc{ExpansionKind: "duplicate", Func: func(*Context, Expression) (string, error) { return "", nil }}
	second := ExpanderFunc{ExpansionKind: "duplicate", Func: func(*Context, Expression) (string, error) { return "", nil }}
	if _, err := New(nil, first, second); err == nil {
		t.Fatal("expected duplicate-kind error")
	}
}

func TestNewRejectsReservedUnknownKind(t *testing.T) {
	t.Parallel()

	expander := ExpanderFunc{
		ExpansionKind: KindUnknown,
		Func:          func(*Context, Expression) (string, error) { return "unexpected", nil },
	}
	if _, err := New(nil, expander); err == nil {
		t.Fatal("expected reserved-kind registration error")
	}
}
