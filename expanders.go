package shellexpand

// DefaultExpanders returns a fresh slice containing every built-in expander.
// Each entry handles exactly one Kind, so callers can select a subset safely.
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
		ReplaceFirstExpander(),
		ReplaceAllExpander(),
		ReplacePrefixExpander(),
		ReplaceSuffixExpander(),
		UppercaseFirstExpander(),
		UppercaseAllExpander(),
		LowercaseFirstExpander(),
		LowercaseAllExpander(),
		IndirectVariableExpander(),
		IndirectPrefixNamesStarExpander(),
		IndirectPrefixNamesAtExpander(),
	}
}

// DefaultComposer returns the shared immutable composer with all built-in
// expanders registered.
func DefaultComposer() *Composer { return defaultComposer }

var defaultComposer = MustNew(DefaultExpanders()...)
