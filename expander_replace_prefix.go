package shellexpand

// ReplacePrefixExpander implements ${name/#pattern/replacement}.
func ReplacePrefixExpander() Expander {
	return replaceExpander(KindReplacePrefix, replacePrefixMode)
}
