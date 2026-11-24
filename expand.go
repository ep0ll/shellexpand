package shellexpand

// Expand uses every built-in expander registered by DefaultComposer.
func Expand(input string, env Env, opts ...Option) (string, error) {
	return DefaultComposer().Expand(input, env, opts...)
}
