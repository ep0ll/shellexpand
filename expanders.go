package shellexpand

// DefaultExpanders returns a fresh slice containing the built-in variable,
// value-operator, length, substring, and pattern-removal expanders.
func DefaultExpanders() []Expander {
	return []Expander{
		DollarVariableExpander(),
		BracedVariableExpander(),
		DefaultIfUnsetOrEmptyExpander(),
		DefaultIfUnsetExpander(),
		AssignIfUnsetOrEmptyExpander(),
		AssignIfUnsetExpander(),
		ErrorIfUnsetOrEmptyExpander(),
		ErrorIfUnsetExpander(),
		AlternateIfSetAndNonEmptyExpander(),
		AlternateIfSetExpander(),
		LengthExpander(),
		SubstringExpander(),
		RemoveShortestPrefixExpander(),
		RemoveLongestPrefixExpander(),
		RemoveShortestSuffixExpander(),
		RemoveLongestSuffixExpander(),
	}
}

// DefaultComposer returns the shared immutable composer with all currently
// implemented built-in expanders registered.
func DefaultComposer() *Composer { return defaultComposer }

var defaultComposer = MustNew(DefaultExpanders()...)
