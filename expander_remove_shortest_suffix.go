package shellexpand

// RemoveShortestSuffixExpander implements ${name%pattern}.
func RemoveShortestSuffixExpander() Expander {
	return trimExpander(KindRemoveShortestSuffix, false, false)
}
