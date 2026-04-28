package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

func TestRandom_RandomBytes_Good(t *testing.T) {
	assert.Len(t, RandomBytes(32), 32)
}

func TestRandom_RandomBytes_Bad(t *testing.T) {
	assert.Panics(t, func() {
		_ = RandomBytes(-1)
	})
}

func TestRandom_RandomBytes_Ugly(t *testing.T) {
	assert.Equal(t, []byte{}, RandomBytes(0))
}

func TestRandom_RandomString_Good(t *testing.T) {
	token := RandomString(8)
	decoded := HexDecode(token)

	assert.Len(t, token, 16)
	assert.True(t, decoded.OK)
	assert.Len(t, decoded.Value.([]byte), 8)
}

func TestRandom_RandomString_Bad(t *testing.T) {
	assert.Panics(t, func() {
		_ = RandomString(-1)
	})
}

func TestRandom_RandomString_Ugly(t *testing.T) {
	assert.Equal(t, "", RandomString(0))
}

func TestRandom_RandomInt_Good(t *testing.T) {
	value := RandomInt(5, 6)

	assert.Equal(t, 5, value)
}

func TestRandom_RandomInt_Bad(t *testing.T) {
	assert.Panics(t, func() {
		_ = RandomInt(10, 10)
	})
}

func TestRandom_RandomInt_Ugly(t *testing.T) {
	for i := 0; i < 100; i++ {
		value := RandomInt(-3, 3)
		assert.GreaterOrEqual(t, value, -3)
		assert.Less(t, value, 3)
	}
}

func TestRandom_RandPick_Good(t *testing.T) {
	item := RandPick([]string{"a"})

	assert.Equal(t, "a", item)
}

func TestRandom_RandPick_Bad(t *testing.T) {
	assert.Panics(t, func() {
		_ = RandPick([]string{})
	})
}

func TestRandom_RandPick_Ugly(t *testing.T) {
	items := []int{1, 2, 3}

	for i := 0; i < 100; i++ {
		assert.True(t, SliceContains(items, RandPick(items)))
	}
}

func TestRandom_RandIntn_Good(t *testing.T) {
	value := RandIntn(1)

	assert.Equal(t, 0, value)
}

func TestRandom_RandIntn_Bad(t *testing.T) {
	assert.Panics(t, func() {
		_ = RandIntn(0)
	})
}

func TestRandom_RandIntn_Ugly(t *testing.T) {
	for i := 0; i < 100; i++ {
		value := RandIntn(5)
		assert.GreaterOrEqual(t, value, 0)
		assert.Less(t, value, 5)
	}
}
