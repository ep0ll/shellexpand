package engine

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
)

// Composer is an immutable, concurrency-safe syntax and expander registry.
// Syntax order defines precedence when more than one syntax matches an offset.
type Composer struct {
	syntaxes  []Syntax
	expanders map[Kind]Expander
	config    config
}

// New validates and composes syntaxes and expanders. An empty syntax list is
// valid and produces a composer that copies input unchanged.
func New(syntaxes []Syntax, expanders ...Expander) (*Composer, error) {
	registeredSyntaxes := make([]Syntax, 0, len(syntaxes))
	for _, syntax := range syntaxes {
		if isNilInterface(syntax) {
			return nil, errors.New("shellexpand: nil syntax")
		}
		registeredSyntaxes = append(registeredSyntaxes, syntax)
	}

	registry := make(map[Kind]Expander, len(expanders))
	for _, expander := range expanders {
		if isNilInterface(expander) {
			return nil, errors.New("shellexpand: nil expander")
		}
		kind := expander.Kind()
		if kind == "" {
			return nil, errors.New("shellexpand: expander must declare a non-empty kind")
		}
		if _, exists := registry[kind]; exists {
			return nil, fmt.Errorf("shellexpand: duplicate expander for kind %q", kind)
		}
		registry[kind] = expander
	}
	return &Composer{syntaxes: registeredSyntaxes, expanders: registry, config: defaultConfig()}, nil
}

// MustNew is like New but panics when the syntax or expander registry is invalid.
func MustNew(syntaxes []Syntax, expanders ...Expander) *Composer {
	composer, err := New(syntaxes, expanders...)
	if err != nil {
		panic(err)
	}
	return composer
}

func isNilInterface(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

// Expand applies configured syntaxes and dispatches each recognized expression
// by its exact Kind. Unmatched and unselected source text is preserved verbatim.
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
	var out strings.Builder
	out.Grow(len(input))
	parseOptions := ParseOptions{MaxDepth: cfg.maxDepth}

	for offset := 0; offset < len(input); {
		matched := false
		for _, syntax := range c.syntaxes {
			token, ok, err := syntax.Match(input, offset, parseOptions)
			if err != nil {
				return "", err
			}
			if !ok {
				continue
			}
			if token.End <= offset || token.End > len(input) {
				return "", fmt.Errorf("shellexpand: syntax returned invalid match range [%d:%d] for input length %d", offset, token.End, len(input))
			}
			matched = true
			if token.Expression == nil {
				if token.Replace {
					out.WriteString(token.Text)
				} else {
					out.WriteString(input[offset:token.End])
				}
				offset = token.End
				break
			}

			expression := *token.Expression
			raw := input[offset:token.End]
			if expression.Raw == "" {
				expression.Raw = raw
			} else if expression.Raw != raw {
				return "", fmt.Errorf("shellexpand: syntax expression raw value %q does not match source %q", expression.Raw, raw)
			}
			if expression.Kind == "" {
				return "", fmt.Errorf("shellexpand: syntax returned an expression with an empty kind at byte %d", offset)
			}
			if expression.Kind == KindUnknown {
				if cfg.unsupportedError {
					return "", &UnsupportedExpressionError{Raw: expression.Raw}
				}
				out.WriteString(expression.Raw)
				offset = token.End
				break
			}
			expander, exists := c.expanders[expression.Kind]
			if !exists {
				if cfg.unselectedError {
					return "", &UnselectedExpanderError{Kind: expression.Kind, Raw: expression.Raw}
				}
				out.WriteString(expression.Raw)
				offset = token.End
				break
			}
			ctx := &Context{composer: c, env: env, config: cfg, depth: depth}
			value, err := expander.Expand(ctx, expression)
			if err != nil {
				return "", fmt.Errorf("shellexpand: expand %s: %w", expression.Raw, err)
			}
			out.WriteString(value)
			offset = token.End
			break
		}
		if matched {
			continue
		}
		_, size := utf8.DecodeRuneInString(input[offset:])
		if size == 0 {
			size = 1
		}
		out.WriteString(input[offset : offset+size])
		offset += size
	}
	return out.String(), nil
}

// Option configures one expansion call.
type Option func(*config)

// WithStrictUnset makes value lookups of unset variables return an error.
func WithStrictUnset() Option { return func(cfg *config) { cfg.strictUnset = true } }

// WithMaxDepth sets the maximum recursive expansion depth. It must be positive.
func WithMaxDepth(depth int) Option { return func(cfg *config) { cfg.maxDepth = depth } }

// WithUnsupportedError errors when a syntax explicitly returns KindUnknown.
func WithUnsupportedError() Option { return func(cfg *config) { cfg.unsupportedError = true } }

// WithUnselectedError errors when an expression's kind has no registered expander.
func WithUnselectedError() Option { return func(cfg *config) { cfg.unselectedError = true } }
