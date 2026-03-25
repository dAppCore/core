package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- ID ---

func TestUtils_ID_Good(t *testing.T) {
	id := ID()
	assert.True(t, HasPrefix(id, "id-"))
	assert.True(t, len(id) > 5, "ID should have counter + random suffix")
}

func TestUtils_ID_Good_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := ID()
		assert.False(t, seen[id], "ID collision: %s", id)
		seen[id] = true
	}
}

func TestUtils_ID_Ugly_CounterMonotonic(t *testing.T) {
	// IDs should contain increasing counter values
	id1 := ID()
	id2 := ID()
	// Both should start with "id-" and have different counter parts
	assert.NotEqual(t, id1, id2)
	assert.True(t, HasPrefix(id1, "id-"))
	assert.True(t, HasPrefix(id2, "id-"))
}

// --- ValidateName ---

func TestUtils_ValidateName_Good(t *testing.T) {
	r := ValidateName("brain")
	assert.True(t, r.OK)
	assert.Equal(t, "brain", r.Value)
}

func TestUtils_ValidateName_Good_WithDots(t *testing.T) {
	r := ValidateName("process.run")
	assert.True(t, r.OK, "dots in names are valid — used for action namespacing")
}

func TestUtils_ValidateName_Bad_Empty(t *testing.T) {
	r := ValidateName("")
	assert.False(t, r.OK)
}

func TestUtils_ValidateName_Bad_Dot(t *testing.T) {
	r := ValidateName(".")
	assert.False(t, r.OK)
}

func TestUtils_ValidateName_Bad_DotDot(t *testing.T) {
	r := ValidateName("..")
	assert.False(t, r.OK)
}

func TestUtils_ValidateName_Bad_Slash(t *testing.T) {
	r := ValidateName("../escape")
	assert.False(t, r.OK)
}

func TestUtils_ValidateName_Ugly_Backslash(t *testing.T) {
	r := ValidateName("windows\\path")
	assert.False(t, r.OK)
}

// --- SanitisePath ---

func TestUtils_SanitisePath_Good(t *testing.T) {
	assert.Equal(t, "file.txt", SanitisePath("/some/path/file.txt"))
}

func TestUtils_SanitisePath_Bad_Empty(t *testing.T) {
	assert.Equal(t, "invalid", SanitisePath(""))
}

func TestUtils_SanitisePath_Bad_DotDot(t *testing.T) {
	assert.Equal(t, "invalid", SanitisePath(".."))
}

func TestUtils_SanitisePath_Ugly_Traversal(t *testing.T) {
	// PathBase extracts "passwd" — the traversal is stripped
	assert.Equal(t, "passwd", SanitisePath("../../etc/passwd"))
}

// --- FilterArgs ---

func TestUtils_FilterArgs_Good(t *testing.T) {
	args := []string{"deploy", "", "to", "-test.v", "homelab", "-test.paniconexit0"}
	clean := FilterArgs(args)
	assert.Equal(t, []string{"deploy", "to", "homelab"}, clean)
}

func TestUtils_FilterArgs_Empty_Good(t *testing.T) {
	clean := FilterArgs(nil)
	assert.Nil(t, clean)
}

// --- ParseFlag ---

func TestUtils_ParseFlag_ShortValid_Good(t *testing.T) {
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

func TestUtils_ParseFlag_ShortInvalid_Bad(t *testing.T) {
	// Multiple chars with single dash — invalid
	_, _, ok := ParseFlag("-verbose")
	assert.False(t, ok)

	_, _, ok = ParseFlag("-port")
	assert.False(t, ok)
}

func TestUtils_ParseFlag_LongValid_Good(t *testing.T) {
	k, v, ok := ParseFlag("--verbose")
	assert.True(t, ok)
	assert.Equal(t, "verbose", k)
	assert.Equal(t, "", v)

	k, v, ok = ParseFlag("--port=8080")
	assert.True(t, ok)
	assert.Equal(t, "port", k)
	assert.Equal(t, "8080", v)
}

func TestUtils_ParseFlag_LongInvalid_Bad(t *testing.T) {
	// Single char with double dash — invalid
	_, _, ok := ParseFlag("--v")
	assert.False(t, ok)
}

func TestUtils_ParseFlag_NotAFlag_Bad(t *testing.T) {
	_, _, ok := ParseFlag("hello")
	assert.False(t, ok)

	_, _, ok = ParseFlag("")
	assert.False(t, ok)
}

// --- IsFlag ---

func TestUtils_IsFlag_Good(t *testing.T) {
	assert.True(t, IsFlag("-v"))
	assert.True(t, IsFlag("--verbose"))
	assert.True(t, IsFlag("-"))
}

func TestUtils_IsFlag_Bad(t *testing.T) {
	assert.False(t, IsFlag("hello"))
	assert.False(t, IsFlag(""))
}

// --- Arg ---

func TestUtils_Arg_String_Good(t *testing.T) {
	r := Arg(0, "hello", 42, true)
	assert.True(t, r.OK)
	assert.Equal(t, "hello", r.Value)
}

func TestUtils_Arg_Int_Good(t *testing.T) {
	r := Arg(1, "hello", 42, true)
	assert.True(t, r.OK)
	assert.Equal(t, 42, r.Value)
}

func TestUtils_Arg_Bool_Good(t *testing.T) {
	r := Arg(2, "hello", 42, true)
	assert.True(t, r.OK)
	assert.Equal(t, true, r.Value)
}

func TestUtils_Arg_UnsupportedType_Good(t *testing.T) {
	r := Arg(0, 3.14)
	assert.True(t, r.OK)
	assert.Equal(t, 3.14, r.Value)
}

func TestUtils_Arg_OutOfBounds_Bad(t *testing.T) {
	r := Arg(5, "only", "two")
	assert.False(t, r.OK)
	assert.Nil(t, r.Value)
}

func TestUtils_Arg_NoArgs_Bad(t *testing.T) {
	r := Arg(0)
	assert.False(t, r.OK)
	assert.Nil(t, r.Value)
}

func TestUtils_Arg_ErrorDetection_Good(t *testing.T) {
	err := NewError("fail")
	r := Arg(0, err)
	assert.True(t, r.OK)
	assert.Equal(t, err, r.Value)
}

// --- ArgString ---

func TestUtils_ArgString_Good(t *testing.T) {
	assert.Equal(t, "hello", ArgString(0, "hello", 42))
	assert.Equal(t, "world", ArgString(1, "hello", "world"))
}

func TestUtils_ArgString_WrongType_Bad(t *testing.T) {
	assert.Equal(t, "", ArgString(0, 42))
}

func TestUtils_ArgString_OutOfBounds_Bad(t *testing.T) {
	assert.Equal(t, "", ArgString(3, "only"))
}

// --- ArgInt ---

func TestUtils_ArgInt_Good(t *testing.T) {
	assert.Equal(t, 42, ArgInt(0, 42, "hello"))
	assert.Equal(t, 99, ArgInt(1, 0, 99))
}

func TestUtils_ArgInt_WrongType_Bad(t *testing.T) {
	assert.Equal(t, 0, ArgInt(0, "not an int"))
}

func TestUtils_ArgInt_OutOfBounds_Bad(t *testing.T) {
	assert.Equal(t, 0, ArgInt(5, 1, 2))
}

// --- ArgBool ---

func TestUtils_ArgBool_Good(t *testing.T) {
	assert.Equal(t, true, ArgBool(0, true, "hello"))
	assert.Equal(t, false, ArgBool(1, true, false))
}

func TestUtils_ArgBool_WrongType_Bad(t *testing.T) {
	assert.Equal(t, false, ArgBool(0, "not a bool"))
}

func TestUtils_ArgBool_OutOfBounds_Bad(t *testing.T) {
	assert.Equal(t, false, ArgBool(5, true))
}

// --- Result.Result() ---

func TestUtils_Result_Result_SingleArg_Good(t *testing.T) {
	r := Result{}.Result("value")
	assert.True(t, r.OK)
	assert.Equal(t, "value", r.Value)
}

func TestUtils_Result_Result_NilError_Good(t *testing.T) {
	r := Result{}.Result("value", nil)
	assert.True(t, r.OK)
	assert.Equal(t, "value", r.Value)
}

func TestUtils_Result_Result_WithError_Bad(t *testing.T) {
	err := NewError("fail")
	r := Result{}.Result("value", err)
	assert.False(t, r.OK)
	assert.Equal(t, err, r.Value)
}

func TestUtils_Result_Result_ZeroArgs_Good(t *testing.T) {
	r := Result{"hello", true}
	got := r.Result()
	assert.Equal(t, "hello", got.Value)
	assert.True(t, got.OK)
}

func TestUtils_Result_Result_ZeroArgs_Empty_Good(t *testing.T) {
	r := Result{}
	got := r.Result()
	assert.Nil(t, got.Value)
	assert.False(t, got.OK)
}
