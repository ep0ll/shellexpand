package shellexpand

// Kind identifies one precise surface form of Bash parameter expansion.
type Kind string

const (
	KindDollarVariable            Kind = "dollar_variable"            // $name
	KindBracedVariable            Kind = "braced_variable"            // ${name}
	KindDefaultIfUnsetOrEmpty     Kind = "default_if_unset_or_empty"  // ${name:-word}
	KindDefaultIfUnset            Kind = "default_if_unset"           // ${name-word}
	KindAssignIfUnsetOrEmpty      Kind = "assign_if_unset_or_empty"   // ${name:=word}
	KindAssignIfUnset             Kind = "assign_if_unset"            // ${name=word}
	KindErrorIfUnsetOrEmpty       Kind = "error_if_unset_or_empty"    // ${name:?word}
	KindErrorIfUnset              Kind = "error_if_unset"             // ${name?word}
	KindAlternateIfSetAndNonEmpty Kind = "alternate_if_set_non_empty" // ${name:+word}
	KindAlternateIfSet            Kind = "alternate_if_set"           // ${name+word}
	KindLength                    Kind = "length"                     // ${#name}
	KindSubstring                 Kind = "substring"                  // ${name:offset[:length]}
	KindRemoveShortestPrefix      Kind = "remove_shortest_prefix"     // ${name#pattern}
	KindRemoveLongestPrefix       Kind = "remove_longest_prefix"      // ${name##pattern}
	KindRemoveShortestSuffix      Kind = "remove_shortest_suffix"     // ${name%pattern}
	KindRemoveLongestSuffix       Kind = "remove_longest_suffix"      // ${name%%pattern}
	KindReplaceFirst              Kind = "replace_first"              // ${name/pattern/replacement}
	KindReplaceAll                Kind = "replace_all"                // ${name//pattern/replacement}
	KindReplacePrefix             Kind = "replace_prefix"             // ${name/#pattern/replacement}
	KindReplaceSuffix             Kind = "replace_suffix"              // ${name/%pattern/replacement}
	KindUppercaseFirst            Kind = "uppercase_first"            // ${name^pattern}
	KindUppercaseAll              Kind = "uppercase_all"              // ${name^^pattern}
	KindLowercaseFirst            Kind = "lowercase_first"            // ${name,pattern}
	KindLowercaseAll              Kind = "lowercase_all"              // ${name,,pattern}
	KindIndirectVariable          Kind = "indirect_variable"          // ${!name}
	KindIndirectPrefixNamesStar   Kind = "indirect_prefix_names_star" // ${!prefix*}
	KindIndirectPrefixNamesAt     Kind = "indirect_prefix_names_at"   // ${!prefix@}
	KindUnknown                   Kind = "unknown"
)

// Expression is one complete parameter expression. Raw always contains the
// complete original spelling, including '$' and (for braced expressions) the
// closing brace. Only fields relevant to Kind are populated.
type Expression struct {
	Kind        Kind
	Raw         string
	Name        string
	Word        string
	Pattern     string
	Replacement string
	Offset      int
	Length      int
	HasLength   bool
}

type part struct {
	text       string
	expression *Expression
}
