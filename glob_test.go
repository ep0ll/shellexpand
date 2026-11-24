package shellexpand

import "testing"

func TestGlobPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pattern string
		value   string
		want    bool
	}{
		{"*", "", true},
		{"*", "a/b\nc", true},
		{"a?c", "abc", true},
		{"a?c", "ac", false},
		{"a\\*c", "a*c", true},
		{"[a-c]", "b", true},
		{"[!a-c]", "Z", true},
		{"[!a-c]", "b", false},
		{"[-a]", "-", true},
		{"[a-]", "-", true},
		{"[]a]", "]", true},
		{"[[]", "[", true},
		{"a/b", "a/b", true},
	}
	for _, test := range tests {
		t.Run(test.pattern+"/"+test.value, func(t *testing.T) {
			got, err := globMatch(test.pattern, test.value)
			if err != nil {
				t.Fatalf("globMatch: %v", err)
			}
			if got != test.want {
				t.Errorf("globMatch(%q, %q) = %v, want %v", test.pattern, test.value, got, test.want)
			}
		})
	}
}

func TestGlobRejectsMalformedPatterns(t *testing.T) {
	t.Parallel()
	for _, pattern := range []string{"[", "[]", "[abc", "abc\\", `[z-a]`, `[[:alpha:]]`, `[[.ch.]]`, `[[=a=]]`} {
		t.Run(pattern, func(t *testing.T) {
			if _, err := compileGlob(pattern); err == nil {
				t.Fatalf("compileGlob(%q) expected error", pattern)
			}
		})
	}
}
