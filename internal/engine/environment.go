package engine

// Env provides variable lookup. The boolean distinguishes an unset variable
// from one explicitly set to the empty string.
type Env interface {
	Lookup(name string) (value string, set bool)
}

// MutableEnv is an Env that can assign values.
type MutableEnv interface {
	Env
	Set(name, value string) error
}

// EnumerableEnv is an Env that can enumerate variable names.
type EnumerableEnv interface {
	Env
	Names() []string
}

// LookupFunc adapts an os.LookupEnv-style function to Env.
type LookupFunc func(name string) (value string, set bool)

// Lookup implements Env. A nil function behaves as an empty environment.
func (f LookupFunc) Lookup(name string) (string, bool) {
	if f == nil {
		return "", false
	}
	return f(name)
}
