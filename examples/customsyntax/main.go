// Command customsyntax demonstrates a non-dollar variable expression.
package main

import (
	"fmt"
	"strings"

	"github.com/ep0ll/shellexpand"
)

type placeholder struct{ name string }

const placeholderKind shellexpand.Kind = "example.double_brace_placeholder"

func main() {
	syntax := shellexpand.SyntaxFunc(func(input string, offset int, _ shellexpand.ParseOptions) (shellexpand.SyntaxMatch, bool, error) {
		if !strings.HasPrefix(input[offset:], "{{") {
			return shellexpand.SyntaxMatch{}, false, nil
		}
		close := strings.Index(input[offset+2:], "}}")
		if close < 0 {
			return shellexpand.SyntaxMatch{}, false, fmt.Errorf("unterminated placeholder at byte %d", offset)
		}
		end := offset + 2 + close + 2
		expression := shellexpand.Expression{
			Kind: placeholderKind,
			Data: placeholder{name: input[offset+2 : end-2]},
		}
		return shellexpand.SyntaxMatch{End: end, Expression: &expression}, true, nil
	})

	expander := shellexpand.ExpanderFunc{
		ExpansionKind: placeholderKind,
		Func: func(ctx *shellexpand.Context, expression shellexpand.Expression) (string, error) {
			token, ok := expression.Data.(placeholder)
			if !ok {
				return "", fmt.Errorf("unexpected payload %T", expression.Data)
			}
			return ctx.Value(token.name)
		},
	}

	composer := shellexpand.MustNewWithSyntax([]shellexpand.Syntax{syntax}, expander)
	got, err := composer.Expand("image={{IMAGE}}", shellexpand.NewMapEnv(map[string]string{"IMAGE": "alpine"}))
	if err != nil {
		panic(err)
	}
	fmt.Println(got)
}
