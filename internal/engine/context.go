package engine

import (
	"errors"
	"fmt"
)

// Context supplies environment access and recursive expansion to an Expander.
// Its fields are private so an expansion cannot replace the active engine state.
type Context struct {
	composer *Composer
	env      Env
	config   config
	depth    int
}

// Lookup returns a variable's value and whether it is set.
func (c *Context) Lookup(name string) (string, bool) {
	if c == nil || isNilInterface(c.env) {
		return "", false
	}
	return c.env.Lookup(name)
}

// Value returns a variable value and applies WithStrictUnset when enabled.
func (c *Context) Value(name string) (string, error) {
	value, set := c.Lookup(name)
	if !set && c.config.strictUnset {
		return "", &UnsetVariableError{Name: name, Message: "not set"}
	}
	return value, nil
}

// Assign sets a variable, returning ErrReadOnlyEnvironment when unsupported.
func (c *Context) Assign(name, value string) error {
	mutable, ok := c.env.(MutableEnv)
	if !ok || isNilInterface(mutable) {
		return fmt.Errorf("shellexpand: %w: %s", ErrReadOnlyEnvironment, name)
	}
	return mutable.Set(name, value)
}

// Names returns variable names when the environment supports enumeration.
func (c *Context) Names() ([]string, bool) {
	enumerable, ok := c.env.(EnumerableEnv)
	if !ok || isNilInterface(enumerable) {
		return nil, false
	}
	return enumerable.Names(), true
}

// ExpandWord recursively expands a word with the same syntax, environment,
// composer, and options. Custom expanders can use it for nested expressions.
func (c *Context) ExpandWord(word string) (string, error) {
	if c == nil || c.composer == nil {
		return "", errors.New("shellexpand: invalid expansion context")
	}
	return c.composer.expand(word, c.env, c.config, c.depth+1)
}

type config struct {
	strictUnset      bool
	maxDepth         int
	unsupportedError bool
	unselectedError  bool
}

func defaultConfig() config { return config{maxDepth: 1000} }
