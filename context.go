package shellexpand

import (
	"errors"
	"fmt"
)

// Context is the evaluation context passed to an Expander implementation.
// Its fields are intentionally private; use the methods below to access the
// environment or recursively expand a word.
type Context struct {
	composer *Composer
	env      Env
	config   config
	depth    int
}

// Lookup returns a variable's value and whether it is set.
func (c *Context) Lookup(name string) (string, bool) {
	if c == nil || c.env == nil {
		return "", false
	}
	return c.env.Lookup(name)
}

// Value returns a variable's value. WithStrictUnset is respected.
func (c *Context) Value(name string) (string, error) {
	value, set := c.Lookup(name)
	if !set && c.config.strictUnset {
		return "", &UnsetVariableError{Name: name, Message: "not set"}
	}
	return value, nil
}

// Assign sets a variable. It returns ErrReadOnlyEnvironment when env does not
// implement MutableEnv.
func (c *Context) Assign(name, value string) error {
	mutable, ok := c.env.(MutableEnv)
	if !ok || mutable == nil {
		return fmt.Errorf("shellexpand: %w: %s", ErrReadOnlyEnvironment, name)
	}
	return mutable.Set(name, value)
}

// Names returns all variable names when env implements EnumerableEnv.
func (c *Context) Names() ([]string, bool) {
	enumerable, ok := c.env.(EnumerableEnv)
	if !ok || enumerable == nil {
		return nil, false
	}
	return enumerable.Names(), true
}

// ExpandWord recursively expands a word using the same composer, environment,
// and options. It is primarily useful to custom Expander implementations.
func (c *Context) ExpandWord(word string) (string, error) {
	if c == nil || c.composer == nil {
		return "", errors.New("shellexpand: invalid expansion context")
	}
	return c.composer.expand(word, c.env, c.config, c.depth+1)
}

// UnsetVariableError describes a required parameter that is unset or empty.
type UnsetVariableError struct {
	Name    string
	Message string
}

func (e *UnsetVariableError) Error() string {
	if e == nil {
		return "shellexpand: unset variable"
	}
	return fmt.Sprintf("$%s: %s", e.Name, e.Message)
}

var (
	// ErrReadOnlyEnvironment is returned by assignment expansions when the
	// supplied Env cannot be mutated.
	ErrReadOnlyEnvironment = errors.New("environment is read-only")
	// ErrVariableNameUnavailable is returned when prefix-indirect expansion
	// requires variable enumeration but Env does not implement EnumerableEnv.
	ErrVariableNameUnavailable = errors.New("environment does not support variable-name enumeration")
)

type config struct {
	strictUnset      bool
	maxDepth         int
	unsupportedError bool
	unselectedError  bool
}

func defaultConfig() config { return config{maxDepth: 1000} }
