package shellexpand

// UppercaseAllExpander implements ${name^^pattern}.
func UppercaseAllExpander() Expander {
	return caseExpander(KindUppercaseAll, true, false)
}
