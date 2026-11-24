package shellexpand

import (
	"slices"
	"testing"
)

func TestProcessEnvLookupAndSet(t *testing.T) {
	const name = "SHELLEXPAND_PROCESS_ENV_TEST"
	t.Setenv(name, "before")

	var env ProcessEnv
	if got, set := env.Lookup(name); !set || got != "before" {
		t.Fatalf("Lookup(%q) = (%q, %v), want (before, true)", name, got, set)
	}
	if err := env.Set(name, "after"); err != nil {
		t.Fatalf("Set(%q): %v", name, err)
	}
	if got, set := env.Lookup(name); !set || got != "after" {
		t.Fatalf("Lookup(%q) = (%q, %v), want (after, true)", name, got, set)
	}
}

func TestProcessEnvNamesAreSortedAndContainCurrentVariables(t *testing.T) {
	const name = "SHELLEXPAND_PROCESS_ENV_NAMES_TEST"
	t.Setenv(name, "value")

	names := (ProcessEnv{}).Names()
	if !slices.IsSorted(names) {
		t.Fatalf("Names() is not sorted: %v", names)
	}
	if !slices.Contains(names, name) {
		t.Fatalf("Names() does not contain %q", name)
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] == names[i] {
			t.Fatalf("Names() contains duplicate %q", names[i])
		}
	}
}
