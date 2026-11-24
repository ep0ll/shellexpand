package shellexpand

import "github.com/ep0ll/shellexpand/internal/environment"

type (
	// MapEnv is an independently owned, concurrency-safe variable store.
	MapEnv = environment.MapEnv
	// ProcessEnv adapts the current process environment.
	ProcessEnv = environment.ProcessEnv
)

// NewMapEnv creates a concurrency-safe in-memory environment.
func NewMapEnv(values map[string]string) *MapEnv { return environment.NewMapEnv(values) }

// NewListEnv parses NAME=value strings into a new MapEnv.
func NewListEnv(pairs ...string) *MapEnv { return environment.NewListEnv(pairs...) }
