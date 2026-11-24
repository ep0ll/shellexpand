package shellexpand

// IndirectPrefixNamesAtExpander implements ${!prefix@}. Like the star form,
// this joins names with spaces because this package returns strings rather
// than Bash's shell-word list.
func IndirectPrefixNamesAtExpander() Expander {
	return prefixNamesExpander(KindIndirectPrefixNamesAt)
}
