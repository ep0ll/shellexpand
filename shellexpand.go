package shellexpand

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Expander evaluates one exact syntactic kind of parameter expansion.
type Expander interface {
	Kind() Kind
	Expand(ctx *Context, expression Expression) (string, error)
}

// ExpanderFunc adapts a function into an Expander.
type ExpanderFunc struct {
	ExpansionKind Kind
	Func          func(*Context, Expression) (string, error)
}

func (f ExpanderFunc) Kind() Kind { return f.ExpansionKind }
func (f ExpanderFunc) Expand(ctx *Context, expression Expression) (string, error) {
	if f.Func == nil {
		return "", errors.New("shellexpand: nil expander function")
	}
	return f.Func(ctx, expression)
}

// Composer is an immutable registry of exact-kind expansion strategies.
type Composer struct {
	expanders map[Kind]Expander
	config    config
}

// New creates a composer. Duplicate kinds and nil expanders are rejected.
func New(expanders ...Expander) (*Composer, error) {
	registry := make(map[Kind]Expander, len(expanders))
	for _, expander := range expanders {
		if isNilExpander(expander) {
			return nil, errors.New("shellexpand: nil expander")
		}
		kind := expander.Kind()
		if kind == KindUnknown || kind == "" {
			return nil, errors.New("shellexpand: expander must declare a supported non-empty kind")
		}
		if _, exists := registry[kind]; exists {
			return nil, fmt.Errorf("shellexpand: duplicate expander for kind %q", kind)
		}
		registry[kind] = expander
	}
	return &Composer{expanders: registry, config: defaultConfig()}, nil
}

func isNilExpander(expander Expander) bool {
	if expander == nil {
		return true
	}
	value := reflect.ValueOf(expander)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// MustNew is like New but panics when the expander set is invalid.
func MustNew(expanders ...Expander) *Composer {
	composer, err := New(expanders...)
	if err != nil {
		panic(err)
	}
	return composer
}

// Expand evaluates input using this composer. Unselected supported
// expressions remain unchanged unless WithUnselectedError is supplied.
func (c *Composer) Expand(input string, env Env, opts ...Option) (string, error) {
	if c == nil {
		return "", errors.New("shellexpand: nil composer")
	}
	cfg := c.config
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.maxDepth <= 0 {
		return "", errors.New("shellexpand: maximum nesting depth must be positive")
	}
	return c.expand(input, env, cfg, 0)
}

func (c *Composer) expand(input string, env Env, cfg config, depth int) (string, error) {
	if depth > cfg.maxDepth {
		return "", fmt.Errorf("shellexpand: expansion nesting exceeds %d levels", cfg.maxDepth)
	}
	parts, err := parse(input, cfg.maxDepth)
	if err != nil {
		return "", err
	}
	ctx := &Context{composer: c, env: env, config: cfg, depth: depth}
	var out strings.Builder
	out.Grow(len(input))
	for _, part := range parts {
		if part.expression == nil {
			out.WriteString(part.text)
			continue
		}
		expression := *part.expression
		if expression.Kind == KindUnknown {
			if cfg.unsupportedError {
				return "", &UnsupportedExpressionError{Raw: expression.Raw}
			}
			out.WriteString(expression.Raw)
			continue
		}
		expander, ok := c.expanders[expression.Kind]
		if !ok {
			if cfg.unselectedError {
				return "", &UnselectedExpanderError{Kind: expression.Kind, Raw: expression.Raw}
			}
			out.WriteString(expression.Raw)
			continue
		}
		value, err := expander.Expand(ctx, expression)
		if err != nil {
			return "", fmt.Errorf("shellexpand: expand %s: %w", expression.Raw, err)
		}
		out.WriteString(value)
	}
	return out.String(), nil
}

// Option configures one expansion call.
type Option func(*config)

// WithStrictUnset makes plain lookups of unset variables return an error.
func WithStrictUnset() Option {
	return func(cfg *config) { cfg.strictUnset = true }
}

// WithMaxDepth sets the maximum recursive word-expansion depth.
func WithMaxDepth(depth int) Option {
	return func(cfg *config) { cfg.maxDepth = depth }
}

// WithUnsupportedError rejects balanced parameter expressions outside the
// supported grammar.
func WithUnsupportedError() Option {
	return func(cfg *config) { cfg.unsupportedError = true }
}

// WithUnselectedError rejects supported expressions without a registered
// expander instead of preserving their original spelling.
func WithUnselectedError() Option {
	return func(cfg *config) { cfg.unselectedError = true }
}

type UnselectedExpanderError struct {
	Kind Kind
	Raw  string
}

func (e *UnselectedExpanderError) Error() string {
	return fmt.Sprintf("no expander registered for %s (%s)", e.Kind, e.Raw)
}

var ErrUnsupportedExpression = errors.New("unsupported parameter expansion")

type UnsupportedExpressionError struct{ Raw string }

func (e *UnsupportedExpressionError) Error() string {
	return fmt.Sprintf("%s %s", ErrUnsupportedExpression, e.Raw)
}
func (e *UnsupportedExpressionError) Unwrap() error { return ErrUnsupportedExpression }
