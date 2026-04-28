package core_test

import (
	. "dappco.re/go"
)

func TestRandom_RandomBytes_Good(t *T) {
	r := RandomBytes(32)
	AssertTrue(t, r.OK)
	AssertLen(t, r.Value.([]byte), 32)
}

func TestRandom_RandomBytes_Bad(t *T) {
	r := RandomBytes(-1)
	AssertFalse(t, r.OK)
}

func TestRandom_RandomBytes_Ugly(t *T) {
	r := RandomBytes(0)
	AssertTrue(t, r.OK)
	AssertEqual(t, []byte{}, r.Value)
}

func TestRandom_RandomString_Good(t *T) {
	r := RandomString(8)
	AssertTrue(t, r.OK)
	token := r.Value.(string)
	decoded := HexDecode(token)

	AssertLen(t, token, 16)
	AssertTrue(t, decoded.OK)
	AssertLen(t, decoded.Value.([]byte), 8)
}

func TestRandom_RandomString_Bad(t *T) {
	r := RandomString(-1)
	AssertFalse(t, r.OK)
}

func TestRandom_RandomString_Ugly(t *T) {
	r := RandomString(0)
	AssertTrue(t, r.OK)
	AssertEqual(t, "", r.Value)
}

func TestRandom_RandomInt_Good(t *T) {
	r := RandomInt(5, 6)
	AssertTrue(t, r.OK)
	AssertEqual(t, 5, r.Value)
}

func TestRandom_RandomInt_Bad(t *T) {
	r := RandomInt(10, 10)
	AssertFalse(t, r.OK)
}

func TestRandom_RandomInt_Ugly(t *T) {
	for i := 0; i < 100; i++ {
		r := RandomInt(-3, 3)
		RequireTrue(t, r.OK)
		value := r.Value.(int)
		AssertGreaterOrEqual(t, value, -3)
		AssertLess(t, value, 3)
	}
}

func TestRandom_RandPick_Good(t *T) {
	item := RandPick([]string{"a"})

	AssertEqual(t, "a", item)
}

func TestRandom_RandPick_Bad(t *T) {
	AssertPanics(t, func() {
		_ = RandPick([]string{})
	})
}

func TestRandom_RandPick_Ugly(t *T) {
	items := []int{1, 2, 3}

	for i := 0; i < 100; i++ {
		AssertTrue(t, SliceContains(items, RandPick(items)))
	}
}

func TestRandom_RandIntn_Good(t *T) {
	value := RandIntn(1)

	AssertEqual(t, 0, value)
}

func TestRandom_RandIntn_Bad(t *T) {
	AssertPanics(t, func() {
		_ = RandIntn(0)
	})
}

func TestRandom_RandIntn_Ugly(t *T) {
	for i := 0; i < 100; i++ {
		value := RandIntn(5)
		AssertGreaterOrEqual(t, value, 0)
		AssertLess(t, value, 5)
	}
}
