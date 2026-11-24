package shellexpand

import "testing"

func TestPatternRemovalExpanders(t *testing.T) {
	env := NewMapEnv(map[string]string{
		"PATH": "/usr/local/bin/tool.tar.gz",
		"FILE": "bash.string.txt",
	})
	tests := []struct{ input, want string }{
		{"${PATH#*/}", "usr/local/bin/tool.tar.gz"},
		{"${PATH##*/}", "tool.tar.gz"},
		{"${FILE%.*}", "bash.string"},
		{"${FILE%%.*}", "bash"},
	}
	for _, tc := range tests {
		got, err := Expand(tc.input, env)
		if err != nil || got != tc.want {
			t.Errorf("Expand(%q) = (%q, %v), want %q", tc.input, got, err, tc.want)
		}
	}
}
