// SPDX-License-Identifier: EUPL-1.2

// Random values for the Core framework.
// Provides cryptographically secure helpers and fast non-crypto selection.

package core

import (
	cryptorand "crypto/rand"
	"math/big"
	fastrand "math/rand/v2"
)

// RandomBytes returns n cryptographically secure random bytes.
// It panics when n is negative.
//
//	token := core.RandomBytes(32)
func RandomBytes(n int) []byte {
	if n < 0 {
		panic("core.RandomBytes: negative length")
	}
	out := make([]byte, n)
	_, _ = cryptorand.Read(out)
	return out
}

// RandomString returns n cryptographically secure random bytes as lowercase hex.
// The returned string has length n*2; it panics when n is negative.
//
//	token := core.RandomString(16)
func RandomString(n int) string {
	return HexEncode(RandomBytes(n))
}

// RandomInt returns a cryptographically secure integer in the half-open range
// [min, max). It panics when max is less than or equal to min.
//
//	delay := core.RandomInt(10, 20)
func RandomInt(min, max int) int {
	span := new(big.Int).Sub(big.NewInt(int64(max)), big.NewInt(int64(min)))
	if span.Sign() <= 0 {
		panic("core.RandomInt: empty range")
	}
	n, err := cryptorand.Int(cryptorand.Reader, span)
	if err != nil {
		panic(err)
	}
	n.Add(n, big.NewInt(int64(min)))
	return int(n.Int64())
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
