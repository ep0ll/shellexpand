package shellexpand

// DefaultExpanders returns a fresh slice containing the built-in variable,
// default, assignment, required, alternate, length, and substring expanders.
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
	}
}

// DefaultComposer returns the shared immutable composer with all currently
// implemented built-in expanders registered.
func DefaultComposer() *Composer { return defaultComposer }

var defaultComposer = MustNew(DefaultExpanders()...)
