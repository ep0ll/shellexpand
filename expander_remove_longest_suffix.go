package shellexpand

// RemoveLongestSuffixExpander implements ${name%%pattern}.
func RemoveLongestSuffixExpander() Expander {
	return trimExpander(KindRemoveLongestSuffix, false, true)
}
