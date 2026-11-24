package shellexpand

import (
	bashexpand "github.com/ep0ll/shellexpand/internal/expanders/bash"
	bashsyntax "github.com/ep0ll/shellexpand/internal/syntax/bash"
)

// BashExpressionData carries the fields parsed by the built-in Bash dialect.
type BashExpressionData = bashsyntax.ExpressionData

const (
	KindUnknown                   Kind = bashsyntax.KindUnknown
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

// BashSyntax returns the built-in Bash parameter-expression recognizer.
func BashSyntax() Syntax { return bashsyntax.NewSyntax() }

// DefaultExpanders returns a new slice with all built-in Bash expanders.
func DefaultExpanders() []Expander {
	return []Expander{
		DollarVariableExpander(), BracedVariableExpander(),
		DefaultIfUnsetOrEmptyExpander(), DefaultIfUnsetExpander(),
		AssignIfUnsetOrEmptyExpander(), AssignIfUnsetExpander(),
		ErrorIfUnsetOrEmptyExpander(), ErrorIfUnsetExpander(),
		AlternateIfSetAndNonEmptyExpander(), AlternateIfSetExpander(),
		LengthExpander(), SubstringExpander(),
		RemoveShortestPrefixExpander(), RemoveLongestPrefixExpander(),
		RemoveShortestSuffixExpander(), RemoveLongestSuffixExpander(),
		ReplaceFirstExpander(), ReplaceAllExpander(), ReplacePrefixExpander(), ReplaceSuffixExpander(),
		UppercaseFirstExpander(), UppercaseAllExpander(), LowercaseFirstExpander(), LowercaseAllExpander(),
		IndirectVariableExpander(), IndirectPrefixNamesStarExpander(), IndirectPrefixNamesAtExpander(),
	}
}

// DefaultComposer returns the shared immutable composer with all Bash expanders.
func DefaultComposer() *Composer { return defaultComposer }

var defaultComposer = MustNew(DefaultExpanders()...)

// DollarVariableExpander expands only the unbraced $name form.
func DollarVariableExpander() Expander { return bashexpand.DollarVariableExpander() }

// BracedVariableExpander expands only the plain ${name} form.
func BracedVariableExpander() Expander { return bashexpand.BracedVariableExpander() }

// DefaultIfUnsetOrEmptyExpander implements ${name:-word}.
func DefaultIfUnsetOrEmptyExpander() Expander { return bashexpand.DefaultIfUnsetOrEmptyExpander() }

// DefaultIfUnsetExpander implements ${name-word}.
func DefaultIfUnsetExpander() Expander { return bashexpand.DefaultIfUnsetExpander() }

// AssignIfUnsetOrEmptyExpander implements ${name:=word}.
func AssignIfUnsetOrEmptyExpander() Expander { return bashexpand.AssignIfUnsetOrEmptyExpander() }

// AssignIfUnsetExpander implements ${name=word}.
func AssignIfUnsetExpander() Expander { return bashexpand.AssignIfUnsetExpander() }

// ErrorIfUnsetOrEmptyExpander implements ${name:?word}.
func ErrorIfUnsetOrEmptyExpander() Expander { return bashexpand.ErrorIfUnsetOrEmptyExpander() }

// ErrorIfUnsetExpander implements ${name?word}.
func ErrorIfUnsetExpander() Expander { return bashexpand.ErrorIfUnsetExpander() }

// AlternateIfSetAndNonEmptyExpander implements ${name:+word}.
func AlternateIfSetAndNonEmptyExpander() Expander {
	return bashexpand.AlternateIfSetAndNonEmptyExpander()
}

// AlternateIfSetExpander implements ${name+word}.
func AlternateIfSetExpander() Expander { return bashexpand.AlternateIfSetExpander() }

// LengthExpander implements ${#name}.
func LengthExpander() Expander { return bashexpand.LengthExpander() }

// SubstringExpander implements ${name:offset[:length]}.
func SubstringExpander() Expander { return bashexpand.SubstringExpander() }

// RemoveShortestPrefixExpander removes the shortest matching prefix.
func RemoveShortestPrefixExpander() Expander { return bashexpand.RemoveShortestPrefixExpander() }

// RemoveLongestPrefixExpander removes the longest matching prefix.
func RemoveLongestPrefixExpander() Expander { return bashexpand.RemoveLongestPrefixExpander() }

// RemoveShortestSuffixExpander removes the shortest matching suffix.
func RemoveShortestSuffixExpander() Expander { return bashexpand.RemoveShortestSuffixExpander() }

// RemoveLongestSuffixExpander removes the longest matching suffix.
func RemoveLongestSuffixExpander() Expander { return bashexpand.RemoveLongestSuffixExpander() }

// ReplaceFirstExpander replaces the first matching pattern.
func ReplaceFirstExpander() Expander { return bashexpand.ReplaceFirstExpander() }

// ReplaceAllExpander replaces every non-overlapping pattern match.
func ReplaceAllExpander() Expander { return bashexpand.ReplaceAllExpander() }

// ReplacePrefixExpander replaces a matching prefix.
func ReplacePrefixExpander() Expander { return bashexpand.ReplacePrefixExpander() }

// ReplaceSuffixExpander replaces a matching suffix.
func ReplaceSuffixExpander() Expander { return bashexpand.ReplaceSuffixExpander() }

// UppercaseFirstExpander uppercases the first matching character.
func UppercaseFirstExpander() Expander { return bashexpand.UppercaseFirstExpander() }

// UppercaseAllExpander uppercases all matching characters.
func UppercaseAllExpander() Expander { return bashexpand.UppercaseAllExpander() }

// LowercaseFirstExpander lowercases the first matching character.
func LowercaseFirstExpander() Expander { return bashexpand.LowercaseFirstExpander() }

// LowercaseAllExpander lowercases all matching characters.
func LowercaseAllExpander() Expander { return bashexpand.LowercaseAllExpander() }

// IndirectVariableExpander resolves a variable name stored in another variable.
func IndirectVariableExpander() Expander { return bashexpand.IndirectVariableExpander() }

// IndirectPrefixNamesStarExpander enumerates names with the given prefix.
func IndirectPrefixNamesStarExpander() Expander {
	return bashexpand.IndirectPrefixNamesStarExpander()
}

// IndirectPrefixNamesAtExpander enumerates names with the given prefix.
func IndirectPrefixNamesAtExpander() Expander { return bashexpand.IndirectPrefixNamesAtExpander() }
