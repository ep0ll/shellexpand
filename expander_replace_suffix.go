package shellexpand

// ReplaceSuffixExpander implements ${name/%pattern/replacement}.
func ReplaceSuffixExpander() Expander {
	return replaceExpander(KindReplaceSuffix, replaceSuffixMode)
}
