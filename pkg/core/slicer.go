// SPDX-License-Identifier: EUPL-1.2

// Generic slice operations for the Core framework.
// Based on leaanthony/slicer, rewritten with Go 1.18+ generics.

package core

// Slicer is a typed slice with common operations.
type Slicer[T comparable] struct {
	items []T
}

// NewSlicer creates an empty Slicer.
func NewSlicer[T comparable](items ...T) *Slicer[T] {
	return &Slicer[T]{items: items}
}

// Add appends values.
func (s *Slicer[T]) Add(values ...T) {
	s.items = append(s.items, values...)
}

// AddUnique appends values only if not already present.
func (s *Slicer[T]) AddUnique(values ...T) {
	for _, v := range values {
		if !s.Contains(v) {
			s.items = append(s.items, v)
		}
	}
}

// Contains returns true if the value is in the slice.
func (s *Slicer[T]) Contains(val T) bool {
	for _, v := range s.items {
		if v == val {
			return true
		}
	}
	return false
}

// Filter returns a new Slicer with elements matching the predicate.
func (s *Slicer[T]) Filter(fn func(T) bool) *Slicer[T] {
	result := &Slicer[T]{}
	for _, v := range s.items {
		if fn(v) {
			result.items = append(result.items, v)
		}
	}
	return result
}

// Each runs a function on every element.
func (s *Slicer[T]) Each(fn func(T)) {
	for _, v := range s.items {
		fn(v)
	}
}

// Remove removes the first occurrence of a value.
func (s *Slicer[T]) Remove(val T) {
	for i, v := range s.items {
		if v == val {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return
		}
	}
}

// Deduplicate removes duplicate values, preserving order.
func (s *Slicer[T]) Deduplicate() {
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
func (s *Slicer[T]) Len() int {
	return len(s.items)
}

// Clear removes all elements.
func (s *Slicer[T]) Clear() {
	s.items = nil
}

// AsSlice returns the underlying slice.
func (s *Slicer[T]) AsSlice() []T {
	return s.items
}
