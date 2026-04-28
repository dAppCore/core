package core_test

import . "dappco.re/go"

// ExampleSetenv sets an environment variable through `Setenv` for process environment
// setup. Process environment mutation goes through core helpers in tests and services.
func ExampleSetenv() {
	Setenv("CORE_EXAMPLE_ENV", "enabled")
	defer UnsetEnv("CORE_EXAMPLE_ENV")

	Println(Env("CORE_EXAMPLE_ENV"))
	// Output: enabled
}

// ExampleUnsetEnv unsets an environment variable through `UnsetEnv` for process
// environment setup. Process environment mutation goes through core helpers in tests and
// services.
func ExampleUnsetEnv() {
	Setenv("CORE_EXAMPLE_ENV", "enabled")
	UnsetEnv("CORE_EXAMPLE_ENV")

	Println(Env("CORE_EXAMPLE_ENV"))
	// Output:
}
