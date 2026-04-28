// SPDX-License-Identifier: EUPL-1.2

// Comparison helpers for the Core framework.

package core

import "cmp"

// Ordered is the canonical ordered constraint for Core generic helpers.
//
//	func min[T core.Ordered](a, b T) T { return core.Min(a, b) }
type Ordered = cmp.Ordered

// Compare returns -1 when a is less than b, 0 when equal, and +1 when greater.
//
//	order := core.Compare("alpha", "beta")
func Compare[T Ordered](a, b T) int {
	return cmp.Compare(a, b)
}
