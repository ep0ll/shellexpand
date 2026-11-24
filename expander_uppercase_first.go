package shellexpand

// UppercaseFirstExpander implements ${name^pattern}.
func UppercaseFirstExpander() Expander {
	return caseExpander(KindUppercaseFirst, true, true)
}
