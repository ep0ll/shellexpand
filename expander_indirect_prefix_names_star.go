package shellexpand

// IndirectPrefixNamesStarExpander implements ${!prefix*}.
func IndirectPrefixNamesStarExpander() Expander {
	return prefixNamesExpander(KindIndirectPrefixNamesStar)
}
