package shellexpand

import (
	"errors"
	"testing"
)

func TestBasicVariableAndDefaultExpansions(t *testing.T) {
	env := NewMapEnv(map[string]string{"NAME":"Ada", "EMPTY":""})
	tests := []struct{ input, want string }{
		{"hello $NAME", "hello Ada"},
		{"hello ${NAME}", "hello Ada"},
		{"${NAME:-fallback}", "Ada"},
		{"${MISSING:-fallback}", "fallback"},
		{"${EMPTY:-fallback}", "fallback"},
		{"${NAME-fallback}", "Ada"},
		{"${MISSING-fallback}", "fallback"},
		{"${EMPTY-fallback}", ""},
		{"${NAME:+alt}", "alt"},
		{"${MISSING:+alt}", ""},
		{"${EMPTY:+alt}", ""},
		{"${NAME+alt}", "alt"},
		{"${MISSING+alt}", ""},
		{"${EMPTY+alt}", "alt"},
		{"${#NAME}", "3"},
	}
	for _, tc := range tests {
		got, err := Expand(tc.input, env)
		if err != nil || got != tc.want {
			t.Errorf("Expand(%q) = (%q, %v), want %q", tc.input, got, err, tc.want)
		}
	}
}

func TestAssignmentAndRequiredExpansion(t *testing.T) {
	for _, tc := range []struct{ input, name, want string }{
		{"${MISSING:=fallback}", "MISSING", "fallback"},
		{"${EMPTY:=fallback}", "EMPTY", "fallback"},
		{"${MISSING=fallback}", "MISSING", "fallback"},
		{"${EMPTY=fallback}", "EMPTY", ""},
	} {
		env := NewMapEnv(map[string]string{"EMPTY":""})
		got, err := Expand(tc.input, env)
		if err != nil || got != tc.want {
			t.Errorf("Expand(%q) = (%q, %v), want %q", tc.input, got, err, tc.want)
		}
		if value, set := env.Lookup(tc.name); !set || value != tc.want {
			t.Errorf("Lookup(%q) = (%q, %v), want (%q, true)", tc.name, value, set, tc.want)
		}
	}

	_, err := Expand("${MISSING:?must be set}", NewMapEnv(nil))
	if err == nil || err.Error() != "shellexpand: expand ${MISSING:?must be set}: $MISSING: must be set" {
		t.Fatalf("required expansion error = %v", err)
	}
	if _, err := Expand("$MISSING", NewMapEnv(nil), WithStrictUnset()); err == nil {
		t.Fatal("strict unset did not report an error")
	}
	if _, err := Expand("${MISSING=fallback}", LookupFunc(func(string)(string,bool){ return "", false })); !errors.Is(err, ErrReadOnlyEnvironment) {
		t.Fatalf("assignment to read-only Env: %v", err)
	}
}

func TestSelectivePlainExpandersPreserveOperators(t *testing.T) {
	composer := MustNew(DollarVariableExpander(), BracedVariableExpander())
	got, err := composer.Expand("$NAME ${NAME} ${NAME-default} ${NAME:-fallback}", NewMapEnv(map[string]string{"NAME":"value"}))
	if err != nil {
		t.Fatal(err)
	}
	want := "value value ${NAME-default} ${NAME:-fallback}"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
