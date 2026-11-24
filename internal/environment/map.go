package environment

import (
	"errors"
	"sort"
	"strings"
	"sync"
)

// MapEnv is an independently owned, concurrency-safe variable store.
type MapEnv struct {
	mu     sync.RWMutex
	values map[string]string
}

// NewMapEnv copies values so callers cannot mutate the environment indirectly.
func NewMapEnv(values map[string]string) *MapEnv {
	copyValues := make(map[string]string, len(values))
	for name, value := range values {
		copyValues[name] = value
	}
	return &MapEnv{values: copyValues}
}

// Lookup returns the value and whether the name is set.
func (e *MapEnv) Lookup(name string) (string, bool) {
	if e == nil {
		return "", false
	}
	e.mu.RLock()
	value, ok := e.values[name]
	e.mu.RUnlock()
	return value, ok
}

// Set assigns a variable.
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

// Names returns sorted variable names for deterministic expansion results.
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

// NewListEnv builds an environment from NAME=value pairs; later values win.
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
