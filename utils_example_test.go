package core_test

import . "dappco.re/go"

func ExampleID_format() {
	id := ID()
	Println(HasPrefix(id, "id-"))
	// Output: true
}

func ExampleValidateName_valid() {
	Println(ValidateName("agent").OK)
	Println(ValidateName("../agent").OK)
	// Output:
	// true
	// false
}

func ExampleSanitisePath_base() {
	Println(SanitisePath("../../etc/passwd"))
	Println(SanitisePath(""))
	// Output:
	// passwd
	// invalid
}

func ExamplePrintln() {
	Println("hello", "codex")
	// Output: hello codex
}

func ExamplePrint() {
	buf := NewBuffer()
	Print(buf, "port=%d", 8080)
	Println(TrimSuffix(buf.String(), "\n"))
	// Output: port=8080
}

func ExampleJoinPath_utils() {
	Println(JoinPath("deploy", "to", "homelab"))
	// Output: deploy/to/homelab
}

func ExampleIsFlag() {
	Println(IsFlag("--verbose"))
	Println(IsFlag("deploy"))
	// Output:
	// true
	// false
}

func ExampleArg() {
	r := Arg(1, "deploy", 8080, true)
	Println(r.Value)
	Println(Arg(9, "deploy").OK)
	// Output:
	// 8080
	// false
}

func ExampleArgString() {
	Println(ArgString(0, "deploy", 8080))
	// Output: deploy
}

func ExampleArgInt() {
	Println(ArgInt(1, "deploy", 8080))
	// Output: 8080
}

func ExampleArgBool() {
	Println(ArgBool(2, "deploy", 8080, true))
	// Output: true
}

func ExampleFilterArgs() {
	Println(FilterArgs([]string{"deploy", "-test.v", "", "--target=homelab"}))
	// Output: [deploy --target=homelab]
}

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
