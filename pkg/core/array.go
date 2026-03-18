// SPDX-License-Identifier: EUPL-1.2

// Generic slice operations for the Core framework.
// Based on leaanthony/slicer, rewritten with Go 1.18+ generics.

package core

// Array is a typed slice with common operations.
type Array[T comparable] struct {
	items []T
}

// NewArray creates an empty Array.
func NewArray[T comparable](items ...T) *Array[T] {
	return &Array[T]{items: items}
}

// Add appends values.
func (s *Array[T]) Add(values ...T) {
	s.items = append(s.items, values...)
}

// AddUnique appends values only if not already present.
func (s *Array[T]) AddUnique(values ...T) {
	for _, v := range values {
		if !s.Contains(v) {
			s.items = append(s.items, v)
		}
	}
}

// Contains returns true if the value is in the slice.
func (s *Array[T]) Contains(val T) bool {
	for _, v := range s.items {
		if v == val {
			return true
		}
	}
	return false
}

// Filter returns a new Array with elements matching the predicate.
func (s *Array[T]) Filter(fn func(T) bool) *Array[T] {
	result := &Array[T]{}
	for _, v := range s.items {
		if fn(v) {
			result.items = append(result.items, v)
		}
	}
	return result
}

// Each runs a function on every element.
func (s *Array[T]) Each(fn func(T)) {
	for _, v := range s.items {
		fn(v)
	}
}

// Remove removes the first occurrence of a value.
func (s *Array[T]) Remove(val T) {
	for i, v := range s.items {
		if v == val {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return
		}
	}
}

// Deduplicate removes duplicate values, preserving order.
func (s *Array[T]) Deduplicate() {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(s.items))
	for _, v := range s.items {
		if _, exists := seen[v]; !exists {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	s.items = result
}

// Len returns the number of elements.
func (s *Array[T]) Len() int {
	return len(s.items)
}

// Clear removes all elements.
func (s *Array[T]) Clear() {
	s.items = nil
}

// AsSlice returns the underlying slice.
func (s *Array[T]) AsSlice() []T {
	return s.items
}
