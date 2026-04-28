package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

func TestMath_Min_Good(t *testing.T) {
	assert.Equal(t, 3, Min(3, 7))
}

func TestMath_Min_Bad(t *testing.T) {
	assert.Equal(t, 3, Min(3, 3))
}

func TestMath_Min_Ugly(t *testing.T) {
	assert.Equal(t, "alpha", Min("beta", "alpha"))
}

func TestMath_Max_Good(t *testing.T) {
	assert.Equal(t, 7, Max(3, 7))
}

func TestMath_Max_Bad(t *testing.T) {
	assert.Equal(t, 3, Max(3, 3))
}

func TestMath_Max_Ugly(t *testing.T) {
	assert.Equal(t, "beta", Max("beta", "alpha"))
}

func TestMath_Abs_Good(t *testing.T) {
	assert.Equal(t, 42, Abs(-42))
}

func TestMath_Abs_Bad(t *testing.T) {
	assert.Equal(t, 0, Abs(0))
}

func TestMath_Abs_Ugly(t *testing.T) {
	assert.Equal(t, float32(3.5), Abs(float32(-3.5)))
}

func TestMath_Pow_Good(t *testing.T) {
	assert.Equal(t, 16.0, Pow(4, 2))
}

func TestMath_Pow_Bad(t *testing.T) {
	assert.Equal(t, 1.0, Pow(4, 0))
}

func TestMath_Pow_Ugly(t *testing.T) {
	assert.InDelta(t, 3.0, Pow(9, 0.5), 0.000001)
}

func TestMath_Floor_Good(t *testing.T) {
	assert.Equal(t, 3.0, Floor(3.7))
}

func TestMath_Floor_Bad(t *testing.T) {
	assert.Equal(t, -4.0, Floor(-3.1))
}

func TestMath_Floor_Ugly(t *testing.T) {
	assert.Equal(t, 3.0, Floor(3))
}

func TestMath_Ceil_Good(t *testing.T) {
	assert.Equal(t, 4.0, Ceil(3.1))
}

func TestMath_Ceil_Bad(t *testing.T) {
	assert.Equal(t, -3.0, Ceil(-3.7))
}

func TestMath_Ceil_Ugly(t *testing.T) {
	assert.Equal(t, 3.0, Ceil(3))
}

func TestMath_Round_Good(t *testing.T) {
	assert.Equal(t, 4.0, Round(3.5))
}

func TestMath_Round_Bad(t *testing.T) {
	assert.Equal(t, -4.0, Round(-3.5))
}

func TestMath_Round_Ugly(t *testing.T) {
	assert.Equal(t, 3.0, Round(3.49))
}
