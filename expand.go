package shellexpand

// Expand evaluates input using all built-in Bash parameter expanders.
func Expand(input string, env Env, opts ...Option) (string, error) {
	return DefaultComposer().Expand(input, env, opts...)
}
