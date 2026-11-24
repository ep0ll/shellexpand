package engine

// ParseOptions carries limits shared by syntax implementations.
type ParseOptions struct {
	MaxDepth int
}

// SyntaxMatch is a single match returned by a Syntax. End is an exclusive
// byte offset into the original input. Set Expression to create an expansion.
// For literal transformations, set Replace and provide Text; this supports
// replacing a token with an empty string. Otherwise, the original token is
// copied unchanged.
type SyntaxMatch struct {
	End        int
	Expression *Expression
	Text       string
	Replace    bool
}

// Syntax recognizes a complete expression or an escaped/literal token at the
// supplied byte offset. A syntax must not consume input when it returns false.
// Multiple syntaxes can be composed; the first syntax that matches wins.
type Syntax interface {
	Match(input string, offset int, options ParseOptions) (SyntaxMatch, bool, error)
}

// SyntaxFunc adapts a function to Syntax.
type SyntaxFunc func(input string, offset int, options ParseOptions) (SyntaxMatch, bool, error)

// Match implements Syntax.
func (f SyntaxFunc) Match(input string, offset int, options ParseOptions) (SyntaxMatch, bool, error) {
	if f == nil {
		return SyntaxMatch{}, false, nil
	}
	return f(input, offset, options)
}
