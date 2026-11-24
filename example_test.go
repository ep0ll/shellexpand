package shellexpand_test

import (
	"fmt"

	"github.com/ep0ll/shellexpand"
)

func ExampleComposer_selectiveExpanders() {
	env := shellexpand.NewMapEnv(map[string]string{
		"NAME":  "world",
		"EMPTY": "",
	})
	composer := shellexpand.MustNew(
		shellexpand.DollarVariableExpander(),
		shellexpand.BracedVariableExpander(),
	)

	got, err := composer.Expand("$NAME ${NAME} ${EMPTY:-fallback}", env)
	if err != nil {
		panic(err)
	}
	fmt.Println(got)
	// Output: world world ${EMPTY:-fallback}
}

func ExampleExpand() {
	env := shellexpand.NewMapEnv(map[string]string{
		"IMAGE": "alpine",
		"TAG":   "3.21",
	})
	got, err := shellexpand.Expand("FROM ${IMAGE}:${TAG:-latest}", env)
	if err != nil {
		panic(err)
	}
	fmt.Println(got)
	// Output: FROM alpine:3.21
}
