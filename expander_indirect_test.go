package shellexpand

import "testing"

func TestIndirectVariableAndPrefixNameExpanders(t *testing.T) {
	env := NewMapEnv(map[string]string{
		"TARGET": "VALUE",
		"VALUE":  "resolved",
		"PRE_A":  "a",
		"PRE_B":  "b",
	})
	tests := []struct{ input, want string }{
		{"${!TARGET}", "resolved"},
		{"${!PRE_*}", "PRE_A PRE_B"},
		{"${!PRE_@}", "PRE_A PRE_B"},
	}
	for _, tc := range tests {
		got, err := Expand(tc.input, env)
		if err != nil || got != tc.want {
			t.Errorf("Expand(%q) = (%q, %v), want %q", tc.input, got, err, tc.want)
		}
	}
}
