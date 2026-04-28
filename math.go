// SPDX-License-Identifier: EUPL-1.2

// Math helpers for the Core framework.

package core

import "math"

type signedOrFloat interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}

// Min returns the smaller of a and b.
//
//	low := core.Min(3, 7)
func Min[T Ordered](a, b T) T {
	if Compare(a, b) <= 0 {
		return a
	}
	return b
}

// Max returns the larger of a and b.
//
//	high := core.Max(3, 7)
func Max[T Ordered](a, b T) T {
	if Compare(a, b) >= 0 {
		return a
	}
	return b
}

// Abs returns the absolute value of x.
//
//	distance := core.Abs(-42)
func Abs[T signedOrFloat](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// IsNaN reports whether x is a floating-point NaN.
//
//	if core.IsNaN(value) { return core.E("math", "not a number", nil) }
func IsNaN(x float64) bool {
	return math.IsNaN(x)
}

// Pow returns x raised to the power y.
//
//	squared := core.Pow(4, 2)
func Pow(x, y float64) float64 {
	return math.Pow(x, y)
}

// Floor returns the greatest integer value less than or equal to x.
//
//	n := core.Floor(3.7)
func Floor(x float64) float64 {
	return math.Floor(x)
}

// Ceil returns the least integer value greater than or equal to x.
//
//	n := core.Ceil(3.1)
func Ceil(x float64) float64 {
	return math.Ceil(x)
}

// Round returns the nearest integer, rounding half away from zero.
//
//	n := core.Round(3.5)
func Round(x float64) float64 {
	return math.Round(x)
}
