package shellexpand

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func testEnv() *MapEnv {
	return NewMapEnv(map[string]string{
		"NAME":    "value",
		"EMPTY":   "",
		"PATH":    "/usr/local/bin/tool.tar.gz",
		"TEXT":    "abcABC123ABCabc",
		"UNICODE": "héllo 🦀",
		"TARGET":  "DEST",
		"DEST":    "indirect-value",
		"PRE_A":   "a",
		"PRE_B":   "b",
	})
}

func TestSelectiveCompositionDoesNotMisdispatch(t *testing.T) {
	t.Parallel()
	composer := MustNew(DollarVariableExpander(), BracedVariableExpander())
	env := testEnv()

	input := `$NAME|${NAME}|${NAME-default}|${NAME:-fallback}|${EMPTY-default}|${EMPTY:-fallback}`
	got, err := composer.Expand(input, env)
	if err != nil {
		t.Fatal(err)
	}
	want := `value|value|${NAME-default}|${NAME:-fallback}|${EMPTY-default}|${EMPTY:-fallback}`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	strict, err := composer.Expand(`${NAME-default}`, env, WithUnselectedError())
	if err == nil {
		t.Fatalf("expected unselected-expander error, got %q", strict)
	}
	var selectionErr *UnselectedExpanderError
	if !errors.As(err, &selectionErr) || selectionErr.Kind != KindDefaultIfUnset {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaultAlternateAndAssignmentOperators(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{`${NAME:-fallback}`, `value`},
		{`${MISSING:-fallback}`, `fallback`},
		{`${EMPTY:-fallback}`, `fallback`},
		{`${NAME-fallback}`, `value`},
		{`${MISSING-fallback}`, `fallback`},
		{`${EMPTY-fallback}`, ``},
		{`${NAME:+alt}`, `alt`},
		{`${MISSING:+alt}`, ``},
		{`${EMPTY:+alt}`, ``},
		{`${NAME+alt}`, `alt`},
		{`${MISSING+alt}`, ``},
		{`${EMPTY+alt}`, `alt`},
	}
	for _, test := range tests {
		got, err := Expand(test.input, testEnv())
		if err != nil || got != test.want {
			t.Errorf("Expand(%q) = (%q, %v), want %q", test.input, got, err, test.want)
		}
	}

	for _, test := range []struct {
		input, name, want string
	}{
		{`${MISSING:=made}`, "MISSING", "made"},
		{`${EMPTY:=made}`, "EMPTY", "made"},
		{`${NAME:=made}`, "NAME", "value"},
		{`${MISSING=made}`, "MISSING", "made"},
		{`${EMPTY=made}`, "EMPTY", ""},
	} {
		env := testEnv()
		got, err := Expand(test.input, env)
		if err != nil || got != test.want {
			t.Errorf("Expand(%q) = (%q, %v), want %q", test.input, got, err, test.want)
		}
		value, set := env.Lookup(test.name)
		if !set || value != test.want {
			t.Errorf("after %q, Lookup(%q) = (%q, %v), want %q and set", test.input, test.name, value, set, test.want)
		}
	}
}

func TestRequiredOperators(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
		err   string
	}{
		{`${NAME?required}`, `value`, ``},
		{`${MISSING?required}`, ``, `$MISSING: required`},
		{`${EMPTY?required}`, ``, ``},
		{`${NAME:?required}`, `value`, ``},
		{`${MISSING:?required}`, ``, `$MISSING: required`},
		{`${EMPTY:?required}`, ``, `$EMPTY: required`},
		{`${EMPTY:?}`, ``, `$EMPTY: null or unset`},
	}
	for _, test := range tests {
		got, err := Expand(test.input, testEnv())
		if test.err == "" {
			if err != nil || got != test.want {
				t.Errorf("Expand(%q) = (%q, %v), want %q", test.input, got, err, test.want)
			}
			continue
		}
		if err == nil || err.Error() != "shellexpand: expand "+test.input+": "+test.err {
			t.Errorf("Expand(%q) error = %v, want suffix %q", test.input, err, test.err)
		}
	}
}

func TestLengthAndSubstring(t *testing.T) {
	t.Parallel()
	tests := []struct{ input, want string }{
		{`${#UNICODE}`, `7`},
		{`${UNICODE:1}`, `éllo 🦀`},
		{`${UNICODE: -2}`, ` 🦀`},
		{`${UNICODE:(-2)}`, ` 🦀`},
		{`${UNICODE:1:3}`, `éll`},
		{`${UNICODE:1:-2}`, `éllo`},
		{`${UNICODE::2}`, `hé`},
		{`${UNICODE:99}`, ``},
		{`${UNICODE: -99}`, ``},
	}
	for _, test := range tests {
		got, err := Expand(test.input, testEnv())
		if err != nil || got != test.want {
			t.Errorf("Expand(%q) = (%q, %v), want %q", test.input, got, err, test.want)
		}
	}
	// The substring expander preserves malformed UTF-8 bytes rather than
	// replacing them with RuneError while finding character boundaries.
	env := NewMapEnv(map[string]string{"BAD": "a\xffb"})
	got, err := Expand(`${BAD:1:1}`, env)
	if err != nil || got != "\xff" {
		t.Fatalf("invalid UTF-8 substring = (%q, %v), want raw 0xff byte", got, err)
	}
	if _, err := Expand(`${UNICODE:1:-99}`, testEnv()); !errors.Is(err, ErrInvalidSubstringRange) {
		t.Fatalf("negative substring range error = %v, want ErrInvalidSubstringRange", err)
	}
}

func TestPatternRemovalAndReplacement(t *testing.T) {
	t.Parallel()
	tests := []struct{ input, want string }{
		{`${PATH#*/}`, `usr/local/bin/tool.tar.gz`},
		{`${PATH##*/}`, `tool.tar.gz`},
		{`${PATH%.*}`, `/usr/local/bin/tool.tar`},
		{`${PATH%%.*}`, `/usr/local/bin/tool`},
		{`${TEXT/abc/X}`, `XABC123ABCabc`},
		{`${TEXT//abc/X}`, `XABC123ABCX`},
		{`${TEXT/#abc/X}`, `XABC123ABCabc`},
		{`${TEXT/%abc/X}`, `abcABC123ABCX`},
		{`${TEXT/[a-c]*/X}`, `X`},
		{`${TEXT/[!a-c]*/X}`, `abcX`},
		{`${TEXT/a\/b/X}`, `abcABC123ABCabc`},
	}
	for _, test := range tests {
		got, err := Expand(test.input, testEnv())
		if err != nil || got != test.want {
			t.Errorf("Expand(%q) = (%q, %v), want %q", test.input, got, err, test.want)
		}
	}
	// These cases distinguish shortest from longest matches and exercise the
	// leftmost-longest replacement rule used by Bash parameter substitution.
	matchEnv := NewMapEnv(map[string]string{
		"VALUE": "abcabc", "MIX": "baa", "PATH": "/a/b", "EMPTY": "", "REPL": `c\/d`,
	})
	more := []struct{ input, want string }{
		{`${VALUE#a*}`, `bcabc`},
		{`${VALUE##a*}`, ``},
		{`${VALUE%*c}`, `abcab`},
		{`${VALUE%%*c}`, ``},
		{`${VALUE/a*/X}`, `X`},
		{`${VALUE//b*/X}`, `aX`},
		{`${MIX/a*/X}`, `bX`},
		{`${MIX//a*/X}`, `bX`},
		{`${PATH//\//-}`, `-a-b`},
		{`${VALUE/abc/}`, `abc`},
		{`${VALUE//abc/}`, ``},
		{`${VALUE/a/c\/d}`, `c/dbcabc`},
		{`${VALUE/a/${EMPTY:-c\/d}}`, `c/dbcabc`},
		{`${VALUE/a/$REPL}`, `c\/dbcabc`},
	}
	for _, test := range more {
		got, err := Expand(test.input, matchEnv)
		if err != nil || got != test.want {
			t.Errorf("Expand(%q) = (%q, %v), want %q", test.input, got, err, test.want)
		}
	}

	empty := NewMapEnv(map[string]string{"EMPTY": ""})
	if got, err := Expand(`${EMPTY//*/X}`, empty); err != nil || got != "X" {
		t.Errorf("empty wildcard replacement = (%q, %v), want %q", got, err, "X")
	}
}

func TestReplacementExpandsNestedWords(t *testing.T) {
	t.Parallel()
	got, err := Expand(`${TEXT/$NAME/$DEST}`, testEnv())
	if err != nil {
		t.Fatal(err)
	}
	// The literal pattern "value" does not occur, so there is no change.
	if got != "abcABC123ABCabc" {
		t.Fatalf("got %q", got)
	}
}

func TestCaseModification(t *testing.T) {
	t.Parallel()
	env := NewMapEnv(map[string]string{"MIXED": "--hello WORLD"})
	tests := []struct{ input, want string }{
		{`${MIXED^}`, `--hello WORLD`},
		{`${MIXED^^}`, `--HELLO WORLD`},
		{`${MIXED,}`, `--hello WORLD`},
		{`${MIXED,,}`, `--hello world`},
		{`${MIXED^W}`, `--hello WORLD`},
		{`${MIXED^[a-z]}`, `--Hello WORLD`},
		{`${MIXED^^[a-z]}`, `--HELLO WORLD`},
		{`${MIXED,,[A-Z]}`, `--hello world`},
	}
	for _, test := range tests {
		got, err := Expand(test.input, env)
		if err != nil || got != test.want {
			t.Errorf("Expand(%q) = (%q, %v), want %q", test.input, got, err, test.want)
		}
	}
}

func TestIndirectExpansion(t *testing.T) {
	t.Parallel()
	env := testEnv()
	tests := []struct{ input, want string }{
		{`${!TARGET}`, `indirect-value`},
		{`${!PRE_*}`, `PRE_A PRE_B`},
		{`${!PRE_@}`, `PRE_A PRE_B`},
	}
	for _, test := range tests {
		got, err := Expand(test.input, env)
		if err != nil || got != test.want {
			t.Errorf("Expand(%q) = (%q, %v), want %q", test.input, got, err, test.want)
		}
	}
	readOnly := LookupFunc(func(name string) (string, bool) { return "DEST", name == "TARGET" })
	if _, err := Expand(`${!TARGET}`, readOnly); err != nil {
		t.Fatal(err)
	}
	if _, err := Expand(`${!PRE_*}`, readOnly); !errors.Is(err, ErrVariableNameUnavailable) {
		t.Fatalf("expected name enumeration error, got %v", err)
	}
}

func TestStrictUnsetAndReadOnlyAssignment(t *testing.T) {
	t.Parallel()
	if _, err := Expand(`$MISSING`, testEnv(), WithStrictUnset()); err == nil {
		t.Fatal("expected strict unset error")
	}
	if _, err := Expand(`${MISSING=word}`, LookupFunc(func(string) (string, bool) { return "", false })); !errors.Is(err, ErrReadOnlyEnvironment) {
		t.Fatalf("expected read-only error, got %v", err)
	}
}

func TestUnsupportedAndUnselectedErrors(t *testing.T) {
	t.Parallel()
	composer := MustNew(DollarVariableExpander())
	if _, err := composer.Expand(`${NAME@Q}`, testEnv(), WithUnsupportedError()); !errors.Is(err, ErrUnsupportedExpression) {
		t.Fatalf("expected unsupported-expression error, got %v", err)
	}
	if _, err := composer.Expand(`${NAME}`, testEnv(), WithUnselectedError()); err == nil {
		t.Fatal("expected unselected-expander error")
	}
}

func TestCustomExpanderAndComposition(t *testing.T) {
	t.Parallel()
	custom := ExpanderFunc{
		ExpansionKind: KindDollarVariable,
		Func: func(_ *Context, e Expression) (string, error) {
			data, ok := e.Data.(BashExpressionData)
			if !ok {
				return "", fmt.Errorf("unexpected Bash expression data %T", e.Data)
			}
			return strings.ToUpper(data.Name), nil
		},
	}
	composer := MustNew(custom, BracedVariableExpander())
	got, err := composer.Expand(`$name:${NAME}`, testEnv())
	if err != nil || got != `NAME:value` {
		t.Fatalf("got (%q, %v)", got, err)
	}
}

func TestDefaultComposerConcurrentUse(t *testing.T) {
	t.Parallel()
	const input = `${NAME} ${EMPTY:-fallback} ${PATH##*/} ${UNICODE:1:2} ${TEXT//abc/x}`
	const want = `value fallback tool.tar.gz él xABC123ABCx`
	var wg sync.WaitGroup
	for range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 20 {
				got, err := Expand(input, testEnv())
				if err != nil || got != want {
					t.Errorf("Expand = (%q, %v), want %q", got, err, want)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestMapEnvConcurrentReadWrite(t *testing.T) {
	t.Parallel()
	env := NewMapEnv(map[string]string{"A": "a"})
	var wg sync.WaitGroup
	for range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 200 {
				_ = env.Set("A", fmt.Sprint(i))
				_, _ = env.Lookup("A")
				_ = env.Names()
			}
		}()
	}
	wg.Wait()
}

func TestNilLookupFuncBehavesAsEmptyEnvironment(t *testing.T) {
	t.Parallel()
	var lookup LookupFunc
	got, err := Expand("$NAME", lookup)
	if err != nil || got != "" {
		t.Fatalf("Expand with nil LookupFunc = (%q, %v), want empty value and no error", got, err)
	}
}
