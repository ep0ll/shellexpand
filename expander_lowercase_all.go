package shellexpand

// LowercaseAllExpander implements ${name,,pattern}.
func LowercaseAllExpander() Expander {
	return caseExpander(KindLowercaseAll, false, false)
}
