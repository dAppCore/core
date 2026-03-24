package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- String Operations ---

func TestHasPrefix_Good(t *testing.T) {
	assert.True(t, HasPrefix("--verbose", "--"))
	assert.True(t, HasPrefix("-v", "-"))
	assert.False(t, HasPrefix("hello", "-"))
}

func TestHasSuffix_Good(t *testing.T) {
	assert.True(t, HasSuffix("test.go", ".go"))
	assert.False(t, HasSuffix("test.go", ".py"))
}

func TestTrimPrefix_Good(t *testing.T) {
	assert.Equal(t, "verbose", TrimPrefix("--verbose", "--"))
	assert.Equal(t, "hello", TrimPrefix("hello", "--"))
}

func TestTrimSuffix_Good(t *testing.T) {
	assert.Equal(t, "test", TrimSuffix("test.go", ".go"))
	assert.Equal(t, "test.go", TrimSuffix("test.go", ".py"))
}

func TestContains_Good(t *testing.T) {
	assert.True(t, Contains("hello world", "world"))
	assert.False(t, Contains("hello world", "mars"))
}

func TestSplit_Good(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, Split("a/b/c", "/"))
}

func TestSplitN_Good(t *testing.T) {
	assert.Equal(t, []string{"key", "value=extra"}, SplitN("key=value=extra", "=", 2))
}

func TestJoin_Good(t *testing.T) {
	assert.Equal(t, "a/b/c", Join("/", "a", "b", "c"))
}

func TestReplace_Good(t *testing.T) {
	assert.Equal(t, "deploy.to.homelab", Replace("deploy/to/homelab", "/", "."))
}

func TestLower_Good(t *testing.T) {
	assert.Equal(t, "hello", Lower("HELLO"))
}

func TestUpper_Good(t *testing.T) {
	assert.Equal(t, "HELLO", Upper("hello"))
}

func TestTrim_Good(t *testing.T) {
	assert.Equal(t, "hello", Trim("  hello  "))
}

func TestRuneCount_Good(t *testing.T) {
	assert.Equal(t, 5, RuneCount("hello"))
	assert.Equal(t, 1, RuneCount("🔥"))
	assert.Equal(t, 0, RuneCount(""))
}
