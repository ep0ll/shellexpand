package shellexpand

// ReplaceAllExpander implements ${name//pattern/replacement}.
func ReplaceAllExpander() Expander {
	return replaceExpander(KindReplaceAll, replaceAllMode)
}
