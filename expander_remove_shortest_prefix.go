package shellexpand

// RemoveShortestPrefixExpander implements ${name#pattern}.
func RemoveShortestPrefixExpander() Expander {
	return trimExpander(KindRemoveShortestPrefix, true, false)
}
