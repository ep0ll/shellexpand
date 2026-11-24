package shellexpand

import (
	"os"
	"slices"
	"strings"
)

// ProcessEnv exposes the current process environment through the Env,
// MutableEnv, and EnumerableEnv interfaces. Its zero value is ready to use.
//
// Assignment expansions evaluated with ProcessEnv call os.Setenv and
// therefore modify the current process environment. Use MapEnv when
// expansion should be isolated from process-global state.
type ProcessEnv struct{}

// Lookup implements Env using os.LookupEnv.
func (ProcessEnv) Lookup(name string) (string, bool) {
	return os.LookupEnv(name)
}

// Set implements MutableEnv using os.Setenv.
func (ProcessEnv) Set(name, value string) error {
	return os.Setenv(name, value)
}

// Names implements EnumerableEnv. The result is sorted and deduplicated so
// prefix-indirect expansion is deterministic even if the platform exposes
// duplicate entries in os.Environ.
func (ProcessEnv) Names() []string {
	pairs := os.Environ()
	names := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		name, _, ok := strings.Cut(pair, "=")
		if ok {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return slices.Compact(names)
}
