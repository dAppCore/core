package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- FilterArgs ---

func TestFilterArgs_Good(t *testing.T) {
	args := []string{"deploy", "", "to", "-test.v", "homelab", "-test.paniconexit0"}
	clean := FilterArgs(args)
	assert.Equal(t, []string{"deploy", "to", "homelab"}, clean)
}

func TestFilterArgs_Empty_Good(t *testing.T) {
	clean := FilterArgs(nil)
	assert.Nil(t, clean)
}

// --- ParseFlag ---

func TestParseFlag_ShortValid_Good(t *testing.T) {
	// Single letter
	k, v, ok := ParseFlag("-v")
	assert.True(t, ok)
	assert.Equal(t, "v", k)
	assert.Equal(t, "", v)

	// Single emoji
	k, v, ok = ParseFlag("-🔥")
	assert.True(t, ok)
	assert.Equal(t, "🔥", k)
	assert.Equal(t, "", v)

	// Short with value
	k, v, ok = ParseFlag("-p=8080")
	assert.True(t, ok)
	assert.Equal(t, "p", k)
	assert.Equal(t, "8080", v)
}

func TestParseFlag_ShortInvalid_Bad(t *testing.T) {
	// Multiple chars with single dash — invalid
	_, _, ok := ParseFlag("-verbose")
	assert.False(t, ok)

	_, _, ok = ParseFlag("-port")
	assert.False(t, ok)
}

func TestParseFlag_LongValid_Good(t *testing.T) {
	k, v, ok := ParseFlag("--verbose")
	assert.True(t, ok)
	assert.Equal(t, "verbose", k)
	assert.Equal(t, "", v)

	k, v, ok = ParseFlag("--port=8080")
	assert.True(t, ok)
	assert.Equal(t, "port", k)
	assert.Equal(t, "8080", v)
}

func TestParseFlag_LongInvalid_Bad(t *testing.T) {
	// Single char with double dash — invalid
	_, _, ok := ParseFlag("--v")
	assert.False(t, ok)
}

func TestParseFlag_NotAFlag_Bad(t *testing.T) {
	_, _, ok := ParseFlag("hello")
	assert.False(t, ok)

	_, _, ok = ParseFlag("")
	assert.False(t, ok)
}
