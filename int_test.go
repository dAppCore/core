package core_test

import (
	. "dappco.re/go"
)

// --- Atoi ---

func TestInt_Atoi_Good(t *T) {
	r := Atoi("42")
	AssertTrue(t, r.OK)
	AssertEqual(t, 42, r.Value)
}

func TestInt_Atoi_Bad(t *T) {
	r := Atoi("not-an-int")
	AssertFalse(t, r.OK)
	_, ok := r.Value.(error)
	AssertTrue(t, ok)
}

func TestInt_Atoi_Ugly(t *T) {
	r := Atoi("999999999999999999999999999999")
	AssertFalse(t, r.OK)
	_, ok := r.Value.(error)
	AssertTrue(t, ok)
}

// --- Itoa ---

func TestInt_Itoa_Good(t *T) {
	AssertEqual(t, "42", Itoa(42))
}

func TestInt_Itoa_Bad(t *T) {
	AssertEqual(t, "-42", Itoa(-42))
}

func TestInt_Itoa_Ugly(t *T) {
	AssertEqual(t, "0", Itoa(0))
}

// --- FormatInt ---

func TestInt_FormatInt_Good(t *T) {
	AssertEqual(t, "ff", FormatInt(255, 16))
}

func TestInt_FormatInt_Bad(t *T) {
	AssertEqual(t, "-ff", FormatInt(-255, 16))
}

func TestInt_FormatInt_Ugly(t *T) {
	AssertEqual(t, "z", FormatInt(35, 36))
	AssertEqual(t, "0", FormatInt(0, 2))
}

// --- ParseInt ---

func TestInt_ParseInt_Good(t *T) {
	r := ParseInt("ff", 16, 64)
	AssertTrue(t, r.OK)
	AssertEqual(t, int64(255), r.Value)
}

func TestInt_ParseInt_Bad(t *T) {
	r := ParseInt("not-an-int", 10, 64)
	AssertFalse(t, r.OK)
	_, ok := r.Value.(error)
	AssertTrue(t, ok)
}

func TestInt_ParseInt_Ugly(t *T) {
	r := ParseInt("255", 10, 8)
	AssertFalse(t, r.OK)
	_, ok := r.Value.(error)
	AssertTrue(t, ok)
}
