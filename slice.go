// SPDX-License-Identifier: EUPL-1.2

// Slice helpers for the Core framework.

package core

import (
	"cmp"
	"slices"
	"sort"
)

// SliceContains reports whether s contains v.
//
//	ok := core.SliceContains([]string{"a", "b"}, "b")
func SliceContains[T comparable](s []T, v T) bool {
	return slices.Contains(s, v)
}

// SliceIndex returns the first index of v in s, or -1 if absent.
//
//	i := core.SliceIndex([]string{"a", "b"}, "b")
func SliceIndex[T comparable](s []T, v T) int {
	return slices.Index(s, v)
}

// SliceSort sorts s in place in ascending order.
//
//	core.SliceSort(scores)
func SliceSort[T cmp.Ordered](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return cmp.Compare(s[i], s[j]) < 0
	})
}

// SliceUniq returns a new slice with duplicate values removed, preserving order.
//
//	names := core.SliceUniq([]string{"a", "b", "a"})
func SliceUniq[T comparable](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	seen := make(map[T]struct{}, len(s))
	out := make([]T, 0, len(s))
	for _, value := range s {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

// SliceReverse reverses s in place.
//
//	core.SliceReverse(items)
func SliceReverse[T any](s []T) {
	slices.Reverse(s)
}
