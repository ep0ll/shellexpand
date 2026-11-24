package bash

import (
	"fmt"

	"github.com/ep0ll/shellexpand/internal/engine"
	bashsyntax "github.com/ep0ll/shellexpand/internal/syntax/bash"
)

type (
	UnsetVariableError = engine.UnsetVariableError
	Context            = engine.Context
	Expression         = engine.Expression
	ExpressionData     = bashsyntax.ExpressionData
	Expander           = engine.Expander
	Kind               = engine.Kind
)

var (
	ErrReadOnlyEnvironment     = engine.ErrReadOnlyEnvironment
	ErrVariableNameUnavailable = engine.ErrVariableNameUnavailable
)

const (
	KindDollarVariable            Kind = bashsyntax.KindDollarVariable
	KindBracedVariable            Kind = bashsyntax.KindBracedVariable
	KindDefaultIfUnsetOrEmpty     Kind = bashsyntax.KindDefaultIfUnsetOrEmpty
	KindDefaultIfUnset            Kind = bashsyntax.KindDefaultIfUnset
	KindAssignIfUnsetOrEmpty      Kind = bashsyntax.KindAssignIfUnsetOrEmpty
	KindAssignIfUnset             Kind = bashsyntax.KindAssignIfUnset
	KindErrorIfUnsetOrEmpty       Kind = bashsyntax.KindErrorIfUnsetOrEmpty
	KindErrorIfUnset              Kind = bashsyntax.KindErrorIfUnset
	KindAlternateIfSetAndNonEmpty Kind = bashsyntax.KindAlternateIfSetAndNonEmpty
	KindAlternateIfSet            Kind = bashsyntax.KindAlternateIfSet
	KindLength                    Kind = bashsyntax.KindLength
	KindSubstring                 Kind = bashsyntax.KindSubstring
	KindRemoveShortestPrefix      Kind = bashsyntax.KindRemoveShortestPrefix
	KindRemoveLongestPrefix       Kind = bashsyntax.KindRemoveLongestPrefix
	KindRemoveShortestSuffix      Kind = bashsyntax.KindRemoveShortestSuffix
	KindRemoveLongestSuffix       Kind = bashsyntax.KindRemoveLongestSuffix
	KindReplaceFirst              Kind = bashsyntax.KindReplaceFirst
	KindReplaceAll                Kind = bashsyntax.KindReplaceAll
	KindReplacePrefix             Kind = bashsyntax.KindReplacePrefix
	KindReplaceSuffix             Kind = bashsyntax.KindReplaceSuffix
	KindUppercaseFirst            Kind = bashsyntax.KindUppercaseFirst
	KindUppercaseAll              Kind = bashsyntax.KindUppercaseAll
	KindLowercaseFirst            Kind = bashsyntax.KindLowercaseFirst
	KindLowercaseAll              Kind = bashsyntax.KindLowercaseAll
	KindIndirectVariable          Kind = bashsyntax.KindIndirectVariable
	KindIndirectPrefixNamesStar   Kind = bashsyntax.KindIndirectPrefixNamesStar
	KindIndirectPrefixNamesAt     Kind = bashsyntax.KindIndirectPrefixNamesAt
)

type builtinExpander struct {
	kind Kind
	fn   func(*Context, Expression) (string, error)
}

func (e builtinExpander) Kind() Kind { return e.kind }
func (e builtinExpander) Expand(ctx *Context, expression Expression) (string, error) {
	return e.fn(ctx, expression)
}

func builtin(kind Kind, fn func(*Context, ExpressionData) (string, error)) Expander {
	return builtinExpander{
		kind: kind,
		fn: func(ctx *Context, expression Expression) (string, error) {
			data, ok := expression.Data.(ExpressionData)
			if !ok {
				return "", fmt.Errorf("shellexpand: invalid Bash expression data for kind %q: %T", kind, expression.Data)
			}
			return fn(ctx, data)
		},
	}
}
