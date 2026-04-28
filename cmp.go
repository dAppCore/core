// SPDX-License-Identifier: EUPL-1.2

// Comparison helpers for the Core framework.

package core

import "cmp"

// Compare returns -1 when a is less than b, 0 when equal, and +1 when greater.
//
//	order := core.Compare("alpha", "beta")
func Compare[T cmp.Ordered](a, b T) int {
	return cmp.Compare(a, b)
}
