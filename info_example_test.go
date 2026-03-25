package core_test

import (
	"fmt"

	. "dappco.re/go/core"
)

func ExampleEnv() {
	fmt.Println(Env("OS"))   // e.g. "darwin"
	fmt.Println(Env("ARCH")) // e.g. "arm64"
}

func ExampleEnvKeys() {
	keys := EnvKeys()
	fmt.Println(len(keys) > 0)
	// Output: true
}
