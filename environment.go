package shellexpand

import (
	"errors"
	"sort"
	"strings"
	"sync"
)

// Env is a source of variable values. Lookup's boolean distinguishes an
// unset variable from one that is set to the empty string.
type Env interface {
	Lookup(name string) (value string, set bool)
}

// MutableEnv is an Env that supports the assignment parameter expansions
// ${name=word} and ${name:=word}.
type MutableEnv interface {
	Env
	Set(name, value string) error
}

// EnumerableEnv is an Env that can list its variable names. It is required by
// ${!prefix*} and ${!prefix@}.
type EnumerableEnv interface {
	Env
	Names() []string
}

// LookupFunc adapts a function, such as os.LookupEnv, to Env.
type LookupFunc func(name string) (value string, set bool)

// Lookup implements Env. A nil LookupFunc behaves as an empty environment.
func (f LookupFunc) Lookup(name string) (string, bool) {
	if f == nil {
		return "", false
	}
	return f(name)
}

// MapEnv is a concurrency-safe in-memory environment. NewMapEnv copies its
// input map, so callers cannot mutate it behind the environment's back.
type MapEnv struct {
	mu     sync.RWMutex
	values map[string]string
}

// NewMapEnv creates an independently owned MapEnv from values.
func NewMapEnv(values map[string]string) *MapEnv {
	copyValues := make(map[string]string, len(values))
	for name, value := range values {
		copyValues[name] = value
	}
	return &MapEnv{values: copyValues}
}

// Lookup implements Env.
func (e *MapEnv) Lookup(name string) (string, bool) {
	if e == nil {
		return "", false
	}
	e.mu.RLock()
	value, ok := e.values[name]
	e.mu.RUnlock()
	return value, ok
}

// Set implements MutableEnv.
func (e *MapEnv) Set(name, value string) error {
	if e == nil {
		return errors.New("shellexpand: cannot assign through a nil MapEnv")
	}
	if name == "" {
		return errors.New("shellexpand: variable name cannot be empty")
	}
	e.mu.Lock()
	if e.values == nil {
		e.values = make(map[string]string)
	}
	e.values[name] = value
	e.mu.Unlock()
	return nil
}

// Names implements EnumerableEnv. The returned names are sorted to make
// prefix-indirect expansions deterministic.
func (e *MapEnv) Names() []string {
	if e == nil {
		return nil
	}
	e.mu.RLock()
	names := make([]string, 0, len(e.values))
	for name := range e.values {
		names = append(names, name)
	}
	e.mu.RUnlock()
	sort.Strings(names)
	return names
}

// NewListEnv creates a MapEnv from NAME=value strings. Entries without an
// equals sign are ignored; later duplicate names replace earlier values.
func NewListEnv(pairs ...string) *MapEnv {
	values := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		name, value, ok := strings.Cut(pair, "=")
		if ok && name != "" {
			values[name] = value
		}
	}
	return NewMapEnv(values)
}
