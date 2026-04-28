package core_test

import (
	. "dappco.re/go"
)

// ExampleSysInfo declares runtime system information through `SysInfo` for runtime
// diagnostics. Runtime and environment facts are exposed as stable diagnostic values.
func ExampleSysInfo() {
	var info *SysInfo
	_ = info
}

// ExampleEnv reads environment access through `Env` for runtime diagnostics. Runtime and
// environment facts are exposed as stable diagnostic values.
func ExampleEnv() {
	Println(Env("OS"))   // e.g. "darwin"
	Println(Env("ARCH")) // e.g. "arm64"
}

// ExampleEnvKeys lists environment keys through `EnvKeys` for runtime diagnostics. Runtime
// and environment facts are exposed as stable diagnostic values.
func ExampleEnvKeys() {
	keys := EnvKeys()
	Println(len(keys) > 0)
	// Output: true
}
