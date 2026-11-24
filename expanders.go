package shellexpand

// DefaultExpanders returns a fresh slice containing all currently implemented
// built-in parameter-value expansion strategies.
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
	}
}

func DefaultComposer() *Composer { return defaultComposer }

var defaultComposer = MustNew(DefaultExpanders()...)
