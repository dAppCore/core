package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- String Operations ---

func TestString_HasPrefix_Good(t *testing.T) {
	assert.True(t, HasPrefix("--verbose", "--"))
	assert.True(t, HasPrefix("-v", "-"))
	assert.False(t, HasPrefix("hello", "-"))
}

func TestString_HasSuffix_Good(t *testing.T) {
	assert.True(t, HasSuffix("test.go", ".go"))
	assert.False(t, HasSuffix("test.go", ".py"))
}

func TestString_TrimPrefix_Good(t *testing.T) {
	assert.Equal(t, "verbose", TrimPrefix("--verbose", "--"))
	assert.Equal(t, "hello", TrimPrefix("hello", "--"))
}

func TestString_TrimSuffix_Good(t *testing.T) {
	assert.Equal(t, "test", TrimSuffix("test.go", ".go"))
	assert.Equal(t, "test.go", TrimSuffix("test.go", ".py"))
}

func TestString_Contains_Good(t *testing.T) {
	assert.True(t, Contains("hello world", "world"))
	assert.False(t, Contains("hello world", "mars"))
}

func TestString_Split_Good(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, Split("a/b/c", "/"))
}

func TestString_SplitN_Good(t *testing.T) {
	assert.Equal(t, []string{"key", "value=extra"}, SplitN("key=value=extra", "=", 2))
}

func TestString_Join_Good(t *testing.T) {
	assert.Equal(t, "a/b/c", Join("/", "a", "b", "c"))
}

func TestString_Replace_Good(t *testing.T) {
	assert.Equal(t, "deploy.to.homelab", Replace("deploy/to/homelab", "/", "."))
}

func TestString_Lower_Good(t *testing.T) {
	assert.Equal(t, "hello", Lower("HELLO"))
}

func TestString_Upper_Good(t *testing.T) {
	assert.Equal(t, "HELLO", Upper("hello"))
}

func TestString_Trim_Good(t *testing.T) {
	assert.Equal(t, "hello", Trim("  hello  "))
}

func TestString_RuneCount_Good(t *testing.T) {
	assert.Equal(t, 5, RuneCount("hello"))
	assert.Equal(t, 1, RuneCount("🔥"))
	assert.Equal(t, 0, RuneCount(""))
}
