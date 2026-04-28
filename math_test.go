package core_test

import (
	. "dappco.re/go/core"
)

func TestMath_Min_Good(t *T) {
	AssertEqual(t, 3, Min(3, 7))
}

func TestMath_Min_Bad(t *T) {
	AssertEqual(t, 3, Min(3, 3))
}

func TestMath_Min_Ugly(t *T) {
	AssertEqual(t, "alpha", Min("beta", "alpha"))
}

func TestMath_Max_Good(t *T) {
	AssertEqual(t, 7, Max(3, 7))
}

func TestMath_Max_Bad(t *T) {
	AssertEqual(t, 3, Max(3, 3))
}

func TestMath_Max_Ugly(t *T) {
	AssertEqual(t, "beta", Max("beta", "alpha"))
}

func TestMath_Abs_Good(t *T) {
	AssertEqual(t, 42, Abs(-42))
}

func TestMath_Abs_Bad(t *T) {
	AssertEqual(t, 0, Abs(0))
}

func TestMath_Abs_Ugly(t *T) {
	AssertEqual(t, float32(3.5), Abs(float32(-3.5)))
}

func TestMath_Pow_Good(t *T) {
	AssertEqual(t, 16.0, Pow(4, 2))
}

func TestMath_Pow_Bad(t *T) {
	AssertEqual(t, 1.0, Pow(4, 0))
}

func TestMath_Pow_Ugly(t *T) {
	AssertInDelta(t, 3.0, Pow(9, 0.5), 0.000001)
}

func TestMath_Floor_Good(t *T) {
	AssertEqual(t, 3.0, Floor(3.7))
}

func TestMath_Floor_Bad(t *T) {
	AssertEqual(t, -4.0, Floor(-3.1))
}

func TestMath_Floor_Ugly(t *T) {
	AssertEqual(t, 3.0, Floor(3))
}

func TestMath_Ceil_Good(t *T) {
	AssertEqual(t, 4.0, Ceil(3.1))
}

func TestMath_Ceil_Bad(t *T) {
	AssertEqual(t, -3.0, Ceil(-3.7))
}

func TestMath_Ceil_Ugly(t *T) {
	AssertEqual(t, 3.0, Ceil(3))
}

func TestMath_Round_Good(t *T) {
	AssertEqual(t, 4.0, Round(3.5))
}

func TestMath_Round_Bad(t *T) {
	AssertEqual(t, -4.0, Round(-3.5))
}

func TestMath_Round_Ugly(t *T) {
	AssertEqual(t, 3.0, Round(3.49))
}
