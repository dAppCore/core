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

// SliceFilter returns a new slice containing only elements for which
// pred returns true. Order is preserved.
//
//	even := core.SliceFilter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
func SliceFilter[T any](s []T, pred func(T) bool) []T {
	if len(s) == 0 {
		return nil
	}
	out := make([]T, 0, len(s))
	for _, v := range s {
		if pred(v) {
			out = append(out, v)
		}
	}
	return out
}

// SliceMap returns a new slice where each element of s has been
// transformed by fn. Note: the name is "SliceMap" (transform a slice),
// not "Map" — core.Map* funcs operate on key-value maps.
//
//	upper := core.SliceMap([]string{"a", "b"}, strings.ToUpper)
//	lengths := core.SliceMap(words, func(w string) int { return len(w) })
func SliceMap[T any, U any](s []T, fn func(T) U) []U {
	if len(s) == 0 {
		return nil
	}
	out := make([]U, len(s))
	for i, v := range s {
		out[i] = fn(v)
	}
	return out
}

// SliceReduce folds s into a single accumulated value, starting from
// init and applying fn to each element.
//
//	sum := core.SliceReduce([]int{1, 2, 3}, 0, func(acc, n int) int { return acc + n })
//	concat := core.SliceReduce([]string{"a","b"}, "", func(acc, s string) string { return acc + s })
func SliceReduce[T any, U any](s []T, init U, fn func(U, T) U) U {
	acc := init
	for _, v := range s {
		acc = fn(acc, v)
	}
	return acc
}

// SliceFlatMap applies fn to each element of s and concatenates the
// resulting slices into one. Useful for "expand each input into 0-N
// outputs" patterns.
//
//	tokens := core.SliceFlatMap(lines, func(l string) []string {
//	    return core.Split(l, " ")
//	})
func SliceFlatMap[T any, U any](s []T, fn func(T) []U) []U {
	if len(s) == 0 {
		return nil
	}
	var out []U
	for _, v := range s {
		out = append(out, fn(v)...)
	}
	return out
}

// SliceTake returns at most the first n elements of s. If n >= len(s),
// returns the whole slice. If n <= 0, returns nil.
//
//	first3 := core.SliceTake(items, 3)
func SliceTake[T any](s []T, n int) []T {
	if n <= 0 {
		return nil
	}
	if n >= len(s) {
		return s
	}
	out := make([]T, n)
	copy(out, s[:n])
	return out
}

// SliceDrop returns s with the first n elements removed. If n >= len(s),
// returns nil. If n <= 0, returns the whole slice.
//
//	rest := core.SliceDrop(items, 1)  // tail
func SliceDrop[T any](s []T, n int) []T {
	if n <= 0 {
		return s
	}
	if n >= len(s) {
		return nil
	}
	out := make([]T, len(s)-n)
	copy(out, s[n:])
	return out
}

// SliceAny reports whether at least one element of s satisfies pred.
//
//	hasNeg := core.SliceAny(nums, func(n int) bool { return n < 0 })
func SliceAny[T any](s []T, pred func(T) bool) bool {
	for _, v := range s {
		if pred(v) {
			return true
		}
	}
	return false
}

// SliceAll reports whether every element of s satisfies pred. Vacuously
// true for empty slices.
//
//	allPositive := core.SliceAll(nums, func(n int) bool { return n > 0 })
func SliceAll[T any](s []T, pred func(T) bool) bool {
	for _, v := range s {
		if !pred(v) {
			return false
		}
	}
	return true
}
