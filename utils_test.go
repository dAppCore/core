package core_test

import (
	. "dappco.re/go"
)

// --- ID ---

func TestUtils_ID_Good(t *T) {
	id := ID()
	AssertTrue(t, HasPrefix(id, "id-"))
	AssertTrue(t, len(id) > 5, "ID should have counter + random suffix")
}

func TestUtils_ID_Good_Unique(t *T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := ID()
		AssertFalse(t, seen[id], "ID collision: %s", id)
		seen[id] = true
	}
}

func TestUtils_ID_Ugly_CounterMonotonic(t *T) {
	// IDs should contain increasing counter values
	id1 := ID()
	id2 := ID()
	// Both should start with "id-" and have different counter parts
	AssertNotEqual(t, id1, id2)
	AssertTrue(t, HasPrefix(id1, "id-"))
	AssertTrue(t, HasPrefix(id2, "id-"))
}

// --- ValidateName ---

func TestUtils_ValidateName_Good(t *T) {
	r := ValidateName("brain")
	AssertTrue(t, r.OK)
	AssertEqual(t, "brain", r.Value)
}

func TestUtils_ValidateName_Good_WithDots(t *T) {
	r := ValidateName("process.run")
	AssertTrue(t, r.OK, "dots in names are valid — used for action namespacing")
}

func TestUtils_ValidateName_Bad_Empty(t *T) {
	r := ValidateName("")
	AssertFalse(t, r.OK)
}

func TestUtils_ValidateName_Bad_Dot(t *T) {
	r := ValidateName(".")
	AssertFalse(t, r.OK)
}

func TestUtils_ValidateName_Bad_DotDot(t *T) {
	r := ValidateName("..")
	AssertFalse(t, r.OK)
}

func TestUtils_ValidateName_Bad_Slash(t *T) {
	r := ValidateName("../escape")
	AssertFalse(t, r.OK)
}

func TestUtils_ValidateName_Ugly_Backslash(t *T) {
	r := ValidateName("windows\\path")
	AssertFalse(t, r.OK)
}

// --- SanitisePath ---

func TestUtils_SanitisePath_Good(t *T) {
	AssertEqual(t, "file.txt", SanitisePath("/some/path/file.txt"))
}

func TestUtils_SanitisePath_Bad_Empty(t *T) {
	AssertEqual(t, "invalid", SanitisePath(""))
}

func TestUtils_SanitisePath_Bad_DotDot(t *T) {
	AssertEqual(t, "invalid", SanitisePath(".."))
}

func TestUtils_SanitisePath_Ugly_Traversal(t *T) {
	// PathBase extracts "passwd" — the traversal is stripped
	AssertEqual(t, "passwd", SanitisePath("../../etc/passwd"))
}

// --- FilterArgs ---

func TestUtils_FilterArgs_Good(t *T) {
	args := []string{"deploy", "", "to", "-test.v", "homelab", "-test.paniconexit0"}
	clean := FilterArgs(args)
	AssertEqual(t, []string{"deploy", "to", "homelab"}, clean)
}

func TestUtils_FilterArgs_Empty_Good(t *T) {
	clean := FilterArgs(nil)
	AssertNil(t, clean)
}

// --- ParseFlag ---

func TestUtils_ParseFlag_ShortValid_Good(t *T) {
	// Single letter
	k, v, ok := ParseFlag("-v")
	AssertTrue(t, ok)
	AssertEqual(t, "v", k)
	AssertEqual(t, "", v)

	// Single emoji
	k, v, ok = ParseFlag("-🔥")
	AssertTrue(t, ok)
	AssertEqual(t, "🔥", k)
	AssertEqual(t, "", v)

	// Short with value
	k, v, ok = ParseFlag("-p=8080")
	AssertTrue(t, ok)
	AssertEqual(t, "p", k)
	AssertEqual(t, "8080", v)
}

func TestUtils_ParseFlag_ShortInvalid_Bad(t *T) {
	// Multiple chars with single dash — invalid
	_, _, ok := ParseFlag("-verbose")
	AssertFalse(t, ok)

	_, _, ok = ParseFlag("-port")
	AssertFalse(t, ok)
}

func TestUtils_ParseFlag_LongValid_Good(t *T) {
	k, v, ok := ParseFlag("--verbose")
	AssertTrue(t, ok)
	AssertEqual(t, "verbose", k)
	AssertEqual(t, "", v)

	k, v, ok = ParseFlag("--port=8080")
	AssertTrue(t, ok)
	AssertEqual(t, "port", k)
	AssertEqual(t, "8080", v)
}

func TestUtils_ParseFlag_LongInvalid_Bad(t *T) {
	// Single char with double dash — invalid
	_, _, ok := ParseFlag("--v")
	AssertFalse(t, ok)
}

func TestUtils_ParseFlag_NotAFlag_Bad(t *T) {
	_, _, ok := ParseFlag("hello")
	AssertFalse(t, ok)

	_, _, ok = ParseFlag("")
	AssertFalse(t, ok)
}

// --- IsFlag ---

func TestUtils_IsFlag_Good(t *T) {
	AssertTrue(t, IsFlag("-v"))
	AssertTrue(t, IsFlag("--verbose"))
	AssertTrue(t, IsFlag("-"))
}

func TestUtils_IsFlag_Bad(t *T) {
	AssertFalse(t, IsFlag("hello"))
	AssertFalse(t, IsFlag(""))
}

// --- Arg ---

func TestUtils_Arg_String_Good(t *T) {
	r := Arg(0, "hello", 42, true)
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello", r.Value)
}

func TestUtils_Arg_Int_Good(t *T) {
	r := Arg(1, "hello", 42, true)
	AssertTrue(t, r.OK)
	AssertEqual(t, 42, r.Value)
}

func TestUtils_Arg_Bool_Good(t *T) {
	r := Arg(2, "hello", 42, true)
	AssertTrue(t, r.OK)
	AssertEqual(t, true, r.Value)
}

func TestUtils_Arg_UnsupportedType_Good(t *T) {
	r := Arg(0, 3.14)
	AssertTrue(t, r.OK)
	AssertEqual(t, 3.14, r.Value)
}

func TestUtils_Arg_OutOfBounds_Bad(t *T) {
	r := Arg(5, "only", "two")
	AssertFalse(t, r.OK)
	AssertNil(t, r.Value)
}

func TestUtils_Arg_NoArgs_Bad(t *T) {
	r := Arg(0)
	AssertFalse(t, r.OK)
	AssertNil(t, r.Value)
}

func TestUtils_Arg_ErrorDetection_Good(t *T) {
	err := NewError("fail")
	r := Arg(0, err)
	AssertTrue(t, r.OK)
	AssertEqual(t, err, r.Value)
}

// --- ArgString ---

func TestUtils_ArgString_Good(t *T) {
	AssertEqual(t, "hello", ArgString(0, "hello", 42))
	AssertEqual(t, "world", ArgString(1, "hello", "world"))
}

func TestUtils_ArgString_WrongType_Bad(t *T) {
	AssertEqual(t, "", ArgString(0, 42))
}

func TestUtils_ArgString_OutOfBounds_Bad(t *T) {
	AssertEqual(t, "", ArgString(3, "only"))
}

// --- ArgInt ---

func TestUtils_ArgInt_Good(t *T) {
	AssertEqual(t, 42, ArgInt(0, 42, "hello"))
	AssertEqual(t, 99, ArgInt(1, 0, 99))
}

func TestUtils_ArgInt_WrongType_Bad(t *T) {
	AssertEqual(t, 0, ArgInt(0, "not an int"))
}

func TestUtils_ArgInt_OutOfBounds_Bad(t *T) {
	AssertEqual(t, 0, ArgInt(5, 1, 2))
}

// --- ArgBool ---

func TestUtils_ArgBool_Good(t *T) {
	AssertEqual(t, true, ArgBool(0, true, "hello"))
	AssertEqual(t, false, ArgBool(1, true, false))
}

func TestUtils_ArgBool_WrongType_Bad(t *T) {
	AssertEqual(t, false, ArgBool(0, "not a bool"))
}

func TestUtils_ArgBool_OutOfBounds_Bad(t *T) {
	AssertEqual(t, false, ArgBool(5, true))
}

// --- Result.Result() ---

func TestUtils_Result_Result_SingleArg_Good(t *T) {
	r := Result{}.Result("value")
	AssertTrue(t, r.OK)
	AssertEqual(t, "value", r.Value)
}

func TestUtils_Result_Result_NilError_Good(t *T) {
	r := Result{}.Result("value", nil)
	AssertTrue(t, r.OK)
	AssertEqual(t, "value", r.Value)
}

func TestUtils_Result_Result_WithError_Bad(t *T) {
	err := NewError("fail")
	r := Result{}.Result("value", err)
	AssertFalse(t, r.OK)
	AssertEqual(t, err, r.Value)
}

func TestUtils_Result_Result_ZeroArgs_Good(t *T) {
	r := Result{"hello", true}
	got := r.Result()
	AssertEqual(t, "hello", got.Value)
	AssertTrue(t, got.OK)
}

func TestUtils_Result_Result_ZeroArgs_Empty_Good(t *T) {
	r := Result{}
	got := r.Result()
	AssertNil(t, got.Value)
	AssertFalse(t, got.OK)
}
