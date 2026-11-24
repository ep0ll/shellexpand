package shellexpand

import "testing"

func FuzzExpand(f *testing.F) {
	for _, seed := range []string{
		"",
		"plain text",
		"$NAME ${NAME} ${NAME:-word}",
		"${NAME-word} ${NAME:=word} ${NAME+alt}",
		"${#NAME} ${NAME: -2} ${NAME:1:2}",
		"${NAME#*.} ${NAME%%.*}",
		"${NAME/a/b} ${NAME//a/b}",
		"${NAME^} ${NAME^^[a-z]} ${NAME,,}",
		"${!NAME} ${!PRE_*}",
		"${MISSING:-${NESTED:-fallback}}",
		`\$NAME $$NAME ${MISSING:-a\}b}`,
		"${NAME",
		"${NAME:invalid}",
		"${NAME/[abc/X}",
	} {
		f.Add(seed)
	}

	env := NewMapEnv(map[string]string{
		"NAME":    "example/value.txt",
		"EMPTY":   "",
		"TARGET":  "NAME",
		"PRE_ONE": "one",
	})
	f.Fuzz(func(t *testing.T, input string) {
		_, _ = Expand(input, env)
		composer := MustNew(DollarVariableExpander(), BracedVariableExpander())
		_, _ = composer.Expand(input, env)
	})
}

func BenchmarkExpand(b *testing.B) {
	env := NewMapEnv(map[string]string{
		"IMAGE": "golang",
		"TAG":   "1.23",
		"PATH":  "/usr/local/go/bin/go.tar.gz",
		"EMPTY": "",
	})
	input := "FROM ${IMAGE}:${TAG:-latest}\nWORKDIR ${PATH%/*}\nLABEL version=${TAG}"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Expand(input, env); err != nil {
			b.Fatal(err)
		}
	}
}
