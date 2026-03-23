package core_test

import (
	"errors"
	"testing"

	. "dappco.re/go/core"
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

// --- IsFlag ---

func TestIsFlag_Good(t *testing.T) {
	assert.True(t, IsFlag("-v"))
	assert.True(t, IsFlag("--verbose"))
	assert.True(t, IsFlag("-"))
}

func TestIsFlag_Bad(t *testing.T) {
	assert.False(t, IsFlag("hello"))
	assert.False(t, IsFlag(""))
}

// --- Arg ---

func TestArg_String_Good(t *testing.T) {
	r := Arg(0, "hello", 42, true)
	assert.True(t, r.OK)
	assert.Equal(t, "hello", r.Value)
}

func TestArg_Int_Good(t *testing.T) {
	r := Arg(1, "hello", 42, true)
	assert.True(t, r.OK)
	assert.Equal(t, 42, r.Value)
}

func TestArg_Bool_Good(t *testing.T) {
	r := Arg(2, "hello", 42, true)
	assert.True(t, r.OK)
	assert.Equal(t, true, r.Value)
}

func TestArg_UnsupportedType_Good(t *testing.T) {
	r := Arg(0, 3.14)
	assert.True(t, r.OK)
	assert.Equal(t, 3.14, r.Value)
}

func TestArg_OutOfBounds_Bad(t *testing.T) {
	r := Arg(5, "only", "two")
	assert.False(t, r.OK)
	assert.Nil(t, r.Value)
}

func TestArg_NoArgs_Bad(t *testing.T) {
	r := Arg(0)
	assert.False(t, r.OK)
	assert.Nil(t, r.Value)
}

func TestArg_ErrorDetection_Good(t *testing.T) {
	err := errors.New("fail")
	r := Arg(0, err)
	assert.True(t, r.OK)
	assert.Equal(t, err, r.Value)
}

// --- ArgString ---

func TestArgString_Good(t *testing.T) {
	assert.Equal(t, "hello", ArgString(0, "hello", 42))
	assert.Equal(t, "world", ArgString(1, "hello", "world"))
}

func TestArgString_WrongType_Bad(t *testing.T) {
	assert.Equal(t, "", ArgString(0, 42))
}

func TestArgString_OutOfBounds_Bad(t *testing.T) {
	assert.Equal(t, "", ArgString(3, "only"))
}

// --- ArgInt ---

func TestArgInt_Good(t *testing.T) {
	assert.Equal(t, 42, ArgInt(0, 42, "hello"))
	assert.Equal(t, 99, ArgInt(1, 0, 99))
}

func TestArgInt_WrongType_Bad(t *testing.T) {
	assert.Equal(t, 0, ArgInt(0, "not an int"))
}

func TestArgInt_OutOfBounds_Bad(t *testing.T) {
	assert.Equal(t, 0, ArgInt(5, 1, 2))
}

// --- ArgBool ---

func TestArgBool_Good(t *testing.T) {
	assert.Equal(t, true, ArgBool(0, true, "hello"))
	assert.Equal(t, false, ArgBool(1, true, false))
}

func TestArgBool_WrongType_Bad(t *testing.T) {
	assert.Equal(t, false, ArgBool(0, "not a bool"))
}

func TestArgBool_OutOfBounds_Bad(t *testing.T) {
	assert.Equal(t, false, ArgBool(5, true))
}

// --- Result.Result() ---

func TestResult_Result_SingleArg_Good(t *testing.T) {
	r := Result{}.Result("value")
	assert.True(t, r.OK)
	assert.Equal(t, "value", r.Value)
}

func TestResult_Result_NilError_Good(t *testing.T) {
	r := Result{}.Result("value", nil)
	assert.True(t, r.OK)
	assert.Equal(t, "value", r.Value)
}

func TestResult_Result_WithError_Bad(t *testing.T) {
	err := errors.New("fail")
	r := Result{}.Result("value", err)
	assert.False(t, r.OK)
	assert.Equal(t, err, r.Value)
}

func TestResult_Result_ZeroArgs_Good(t *testing.T) {
	r := Result{"hello", true}
	got := r.Result()
	assert.Equal(t, "hello", got.Value)
	assert.True(t, got.OK)
}

func TestResult_Result_ZeroArgs_Empty_Good(t *testing.T) {
	r := Result{}
	got := r.Result()
	assert.Nil(t, got.Value)
	assert.False(t, got.OK)
}
