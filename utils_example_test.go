package core_test

import . "dappco.re/go"

// ExampleID_format formats an identifier through `ID` for CLI utility parsing. Small CLI
// argument utilities have predictable string and flag behaviour.
func ExampleID_format() {
	id := ID()
	Println(HasPrefix(id, "id-"))
	// Output: true
}

// ExampleValidateName_valid accepts a valid name through `ValidateName` for CLI utility
// parsing. Small CLI argument utilities have predictable string and flag behaviour.
func ExampleValidateName_valid() {
	Println(ValidateName("agent").OK)
	Println(ValidateName("../agent").OK)
	// Output:
	// true
	// false
}

// ExampleSanitisePath_base sanitises a base path through `SanitisePath` for CLI utility
// parsing. Small CLI argument utilities have predictable string and flag behaviour.
func ExampleSanitisePath_base() {
	Println(SanitisePath("../../etc/passwd"))
	Println(SanitisePath(""))
	// Output:
	// passwd
	// invalid
}

// ExamplePrintln writes a line through `Println` for CLI utility parsing. Small CLI
// argument utilities have predictable string and flag behaviour.
func ExamplePrintln() {
	Println("hello", "codex")
	// Output: hello codex
}

// ExamplePrint writes text through `Print` for CLI utility parsing. Small CLI argument
// utilities have predictable string and flag behaviour.
func ExamplePrint() {
	buf := NewBuffer()
	Print(buf, "port=%d", 8080)
	Println(TrimSuffix(buf.String(), "\n"))
	// Output: port=8080
}

// ExampleJoinPath_utils joins path segments through `JoinPath` for CLI utility parsing.
// Small CLI argument utilities have predictable string and flag behaviour.
func ExampleJoinPath_utils() {
	Println(JoinPath("deploy", "to", "homelab"))
	// Output: deploy/to/homelab
}

// ExampleIsFlag checks flag syntax through `IsFlag` for CLI utility parsing. Small CLI
// argument utilities have predictable string and flag behaviour.
func ExampleIsFlag() {
	Println(IsFlag("--verbose"))
	Println(IsFlag("deploy"))
	// Output:
	// true
	// false
}

// ExampleArg reads a positional argument through `Arg` for CLI utility parsing. Small CLI
// argument utilities have predictable string and flag behaviour.
func ExampleArg() {
	r := Arg(1, "deploy", 8080, true)
	Println(r.Value)
	Println(Arg(9, "deploy").OK)
	// Output:
	// 8080
	// false
}

// ExampleArgString reads a string argument through `ArgString` for CLI utility parsing.
// Small CLI argument utilities have predictable string and flag behaviour.
func ExampleArgString() {
	Println(ArgString(0, "deploy", 8080))
	// Output: deploy
}

// ExampleArgInt reads an integer argument through `ArgInt` for CLI utility parsing. Small
// CLI argument utilities have predictable string and flag behaviour.
func ExampleArgInt() {
	Println(ArgInt(1, "deploy", 8080))
	// Output: 8080
}

// ExampleArgBool reads a boolean argument through `ArgBool` for CLI utility parsing. Small
// CLI argument utilities have predictable string and flag behaviour.
func ExampleArgBool() {
	Println(ArgBool(2, "deploy", 8080, true))
	// Output: true
}

// ExampleFilterArgs filters arguments through `FilterArgs` for CLI utility parsing. Small
// CLI argument utilities have predictable string and flag behaviour.
func ExampleFilterArgs() {
	Println(FilterArgs([]string{"deploy", "-test.v", "", "--target=homelab"}))
	// Output: [deploy --target=homelab]
}

// ExampleParseFlag parses a flag through `ParseFlag` for CLI utility parsing. Small CLI
// argument utilities have predictable string and flag behaviour.
func ExampleParseFlag() {
	key, value, ok := ParseFlag("--target=homelab")
	Println(key)
	Println(value)
	Println(ok)
	// Output:
	// target
	// homelab
	// true
}
