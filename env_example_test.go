package core_test

import . "dappco.re/go"

func ExampleSetenv() {
	Setenv("CORE_EXAMPLE_ENV", "enabled")
	defer UnsetEnv("CORE_EXAMPLE_ENV")

	Println(Env("CORE_EXAMPLE_ENV"))
	// Output: enabled
}

func ExampleUnsetEnv() {
	Setenv("CORE_EXAMPLE_ENV", "enabled")
	UnsetEnv("CORE_EXAMPLE_ENV")

	Println(Env("CORE_EXAMPLE_ENV"))
	// Output:
}
