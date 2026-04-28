// SPDX-License-Identifier: EUPL-1.2

// Random values for the Core framework.
// Provides cryptographically secure helpers and fast non-crypto selection.

package core

import (
	cryptorand "crypto/rand"
	"math/big"
	fastrand "math/rand/v2"
)

// RandomBytes returns n cryptographically secure random bytes wrapped in
// a Result. Returns OK=false when n is negative or the OS entropy source
// fails. Code is "random.length.invalid" or "random.entropy.failed".
//
//	r := core.RandomBytes(32)
//	if !r.OK { return r }
//	token := r.Value.([]byte)
func RandomBytes(n int) Result {
	if n < 0 {
		return Result{Value: NewCode("random.length.invalid", "RandomBytes: negative length"), OK: false}
	}
	out := make([]byte, n)
	if _, err := cryptorand.Read(out); err != nil {
		return Result{Value: WrapCode(err, "random.entropy.failed", "RandomBytes", "OS entropy source failed"), OK: false}
	}
	return Result{Value: out, OK: true}
}

// RandomString returns n cryptographically secure random bytes encoded
// as lowercase hex. The Value is a string of length n*2. Code mirrors
// RandomBytes when the underlying source fails.
//
//	r := core.RandomString(16)
//	if r.OK { token := r.Value.(string) }
func RandomString(n int) Result {
	r := RandomBytes(n)
	if !r.OK {
		return r
	}
	return Result{Value: HexEncode(r.Value.([]byte)), OK: true}
}

// RandomInt returns a cryptographically secure integer in the half-open
// range [min, max). Returns OK=false when max <= min (Code
// "random.range.empty") or when crypto/rand fails (Code
// "random.entropy.failed").
//
//	r := core.RandomInt(10, 20)
//	if r.OK { delay := r.Value.(int) }
func RandomInt(min, max int) Result {
	span := new(big.Int).Sub(big.NewInt(int64(max)), big.NewInt(int64(min)))
	if span.Sign() <= 0 {
		return Result{Value: NewCode("random.range.empty", "RandomInt: empty range"), OK: false}
	}
	n, err := cryptorand.Int(cryptorand.Reader, span)
	if err != nil {
		return Result{Value: WrapCode(err, "random.entropy.failed", "RandomInt", "OS entropy source failed"), OK: false}
	}
	n.Add(n, big.NewInt(int64(min)))
	return Result{Value: int(n.Int64()), OK: true}
}

// RandPick returns a pseudo-random item from items.
// It panics when items is empty; use it for fast non-crypto selection only.
//
//	choice := core.RandPick([]string{"red", "green", "blue"})
func RandPick[T any](items []T) T {
	if len(items) == 0 {
		panic("core.RandPick: empty slice")
	}
	return items[RandIntn(len(items))]
}

// RandIntn returns a pseudo-random integer in [0, n).
// It panics when n is less than or equal to zero; use it for fast non-crypto values.
//
//	index := core.RandIntn(len(items))
func RandIntn(n int) int {
	return fastrand.IntN(n)
}
