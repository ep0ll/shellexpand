package shellexpand

import "testing"

func TestPatternReplacementAndCaseExpanders(t *testing.T) {
	env := NewMapEnv(map[string]string{
		"TEXT": "abcABC123ABCabc",
		"MIX":  "--hello WORLD",
	})
	tests := []struct{ input, want string }{
		{"${TEXT/abc/X}", "XABC123ABCabc"},
		{"${TEXT//abc/X}", "XABC123ABCX"},
		{"${TEXT/#abc/X}", "XABC123ABCabc"},
		{"${TEXT/%abc/X}", "abcABC123ABCX"},
		{"${MIX^}", "--hello WORLD"},
		{"${MIX^^}", "--HELLO WORLD"},
		{"${MIX,}", "--hello WORLD"},
		{"${MIX,,}", "--hello world"},
		{"${MIX^^[a-z]}", "--HELLO WORLD"},
		{"${MIX,,[A-Z]}", "--hello world"},
	}
	for _, tc := range tests {
		got, err := Expand(tc.input, env)
		if err != nil || got != tc.want {
			t.Errorf("Expand(%q) = (%q, %v), want %q", tc.input, got, err, tc.want)
		}
	}
}
