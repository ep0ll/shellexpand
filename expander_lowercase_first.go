package shellexpand

// LowercaseFirstExpander implements ${name,pattern}.
func LowercaseFirstExpander() Expander {
	return caseExpander(KindLowercaseFirst, false, true)
}
