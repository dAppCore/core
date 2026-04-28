package core_test

import (
	"testing"

	. "dappco.re/go/core"
)

// --- String Operations ---

func TestString_HasPrefix_Good(t *testing.T) {
	AssertTrue(t, HasPrefix("--verbose", "--"))
	AssertTrue(t, HasPrefix("-v", "-"))
	AssertFalse(t, HasPrefix("hello", "-"))
}

func TestString_HasSuffix_Good(t *testing.T) {
	AssertTrue(t, HasSuffix("test.go", ".go"))
	AssertFalse(t, HasSuffix("test.go", ".py"))
}

func TestString_TrimPrefix_Good(t *testing.T) {
	AssertEqual(t, "verbose", TrimPrefix("--verbose", "--"))
	AssertEqual(t, "hello", TrimPrefix("hello", "--"))
}

func TestString_TrimSuffix_Good(t *testing.T) {
	AssertEqual(t, "test", TrimSuffix("test.go", ".go"))
	AssertEqual(t, "test.go", TrimSuffix("test.go", ".py"))
}

func TestString_Contains_Good(t *testing.T) {
	AssertTrue(t, Contains("hello world", "world"))
	AssertFalse(t, Contains("hello world", "mars"))
}

func TestString_Split_Good(t *testing.T) {
	AssertEqual(t, []string{"a", "b", "c"}, Split("a/b/c", "/"))
}

func TestString_SplitN_Good(t *testing.T) {
	AssertEqual(t, []string{"key", "value=extra"}, SplitN("key=value=extra", "=", 2))
}

func TestString_Join_Good(t *testing.T) {
	AssertEqual(t, "a/b/c", Join("/", "a", "b", "c"))
}

func TestString_Replace_Good(t *testing.T) {
	AssertEqual(t, "deploy.to.homelab", Replace("deploy/to/homelab", "/", "."))
}

func TestString_Lower_Good(t *testing.T) {
	AssertEqual(t, "hello", Lower("HELLO"))
}

func TestString_Upper_Good(t *testing.T) {
	AssertEqual(t, "HELLO", Upper("hello"))
}

func TestString_Trim_Good(t *testing.T) {
	AssertEqual(t, "hello", Trim("  hello  "))
}

func TestString_RuneCount_Good(t *testing.T) {
	AssertEqual(t, 5, RuneCount("hello"))
	AssertEqual(t, 1, RuneCount("🔥"))
	AssertEqual(t, 0, RuneCount(""))
}
