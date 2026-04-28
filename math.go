// SPDX-License-Identifier: EUPL-1.2

// Math and ordering helpers for the Core framework. SPOR owner for
// the math, math/big, and cmp stdlib packages.

package core

import (
	"cmp"
	"math"
)

type signedOrFloat interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}

// Ordered is the canonical ordered constraint for Core generic helpers.
// Alias of cmp.Ordered so consumers stay on the core surface.
//
//	func sortAgents[T core.Ordered](xs []T) { core.SliceSort(xs) }
type Ordered = cmp.Ordered

// Compare returns -1 when a is less than b, 0 when equal, and +1 when
// greater. Wraps cmp.Compare for the canonical ordering.
//
//	order := core.Compare("alpha", "beta")  // -1
//	order := core.Compare(7, 7)             //  0
func Compare[T Ordered](a, b T) int {
	return cmp.Compare(a, b)
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
