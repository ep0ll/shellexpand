package shellexpand

// RemoveLongestPrefixExpander implements ${name##pattern}.
func RemoveLongestPrefixExpander() Expander {
	return trimExpander(KindRemoveLongestPrefix, true, true)
}
