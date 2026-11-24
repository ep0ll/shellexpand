package shellexpand

// ReplaceFirstExpander implements ${name/pattern/replacement}.
func ReplaceFirstExpander() Expander {
	return replaceExpander(KindReplaceFirst, replaceFirstMode)
}
