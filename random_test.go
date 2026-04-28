package core_test

import (
	"testing"

	. "dappco.re/go/core"
)

func TestRandom_RandomBytes_Good(t *testing.T) {
	AssertLen(t, RandomBytes(32), 32)
}

func TestRandom_RandomBytes_Bad(t *testing.T) {
	AssertPanics(t, func() {
		_ = RandomBytes(-1)
	})
}

func TestRandom_RandomBytes_Ugly(t *testing.T) {
	AssertEqual(t, []byte{}, RandomBytes(0))
}

func TestRandom_RandomString_Good(t *testing.T) {
	token := RandomString(8)
	decoded := HexDecode(token)

	AssertLen(t, token, 16)
	AssertTrue(t, decoded.OK)
	AssertLen(t, decoded.Value.([]byte), 8)
}

func TestRandom_RandomString_Bad(t *testing.T) {
	AssertPanics(t, func() {
		_ = RandomString(-1)
	})
}

func TestRandom_RandomString_Ugly(t *testing.T) {
	AssertEqual(t, "", RandomString(0))
}

func TestRandom_RandomInt_Good(t *testing.T) {
	value := RandomInt(5, 6)

	AssertEqual(t, 5, value)
}

func TestRandom_RandomInt_Bad(t *testing.T) {
	AssertPanics(t, func() {
		_ = RandomInt(10, 10)
	})
}

func TestRandom_RandomInt_Ugly(t *testing.T) {
	for i := 0; i < 100; i++ {
		value := RandomInt(-3, 3)
		AssertGreaterOrEqual(t, value, -3)
		AssertLess(t, value, 3)
	}
}

func TestRandom_RandPick_Good(t *testing.T) {
	item := RandPick([]string{"a"})

	AssertEqual(t, "a", item)
}

func TestRandom_RandPick_Bad(t *testing.T) {
	AssertPanics(t, func() {
		_ = RandPick([]string{})
	})
}

func TestRandom_RandPick_Ugly(t *testing.T) {
	items := []int{1, 2, 3}

	for i := 0; i < 100; i++ {
		AssertTrue(t, SliceContains(items, RandPick(items)))
	}
}

func TestRandom_RandIntn_Good(t *testing.T) {
	value := RandIntn(1)

	AssertEqual(t, 0, value)
}

func TestRandom_RandIntn_Bad(t *testing.T) {
	AssertPanics(t, func() {
		_ = RandIntn(0)
	})
}

func TestRandom_RandIntn_Ugly(t *testing.T) {
	for i := 0; i < 100; i++ {
		value := RandIntn(5)
		AssertGreaterOrEqual(t, value, 0)
		AssertLess(t, value, 5)
	}
}
