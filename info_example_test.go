package core_test

import (
	. "dappco.re/go/core"
)

func ExampleEnv() {
	Println(Env("OS"))   // e.g. "darwin"
	Println(Env("ARCH")) // e.g. "arm64"
}

func ExampleEnvKeys() {
	keys := EnvKeys()
	Println(len(keys) > 0)
	// Output: true
}
