// SPDX-License-Identifier: EUPL-1.2

// Generic slice operations for the Core framework.
// Based on leaanthony/slicer, rewritten with Go 1.18+ generics.

package core

// Array is a typed slice with common operations.
//
//	agents := core.NewArray("codex", "hades")
//	agents.AddUnique("homelab")
type Array[T comparable] struct {
	items []T
}

// NewArray creates an empty Array.
//
//	agents := core.NewArray("codex", "hades")
//	core.Println(agents.Len())
func NewArray[T comparable](items ...T) *Array[T] {
	return &Array[T]{items: items}
}

// Add appends values.
//
//	agents := core.NewArray("codex")
//	agents.Add("hades", "homelab")
func (s *Array[T]) Add(values ...T) {
	s.items = append(s.items, values...)
}

// AddUnique appends values only if not already present.
//
//	agents := core.NewArray("codex")
//	agents.AddUnique("codex", "hades")
func (s *Array[T]) AddUnique(values ...T) {
	for _, v := range values {
		if !s.Contains(v) {
			s.items = append(s.items, v)
		}
	}
}

// Contains returns true if the value is in the slice.
//
//	agents := core.NewArray("codex", "hades")
//	if agents.Contains("hades") {
//	    core.Println("agent present")
//	}
func (s *Array[T]) Contains(val T) bool {
	for _, v := range s.items {
		if v == val {
			return true
		}
	}
	return false
}

// Filter returns a new Array with elements matching the predicate.
//
//	agents := core.NewArray("codex", "hades", "homelab")
//	r := agents.Filter(func(name string) bool { return core.HasPrefix(name, "h") })
//	if !r.OK {
//	    return r
//	}
//	filtered := r.Value.(*core.Array[string])
//	core.Println(filtered.Len())
func (s *Array[T]) Filter(fn func(T) bool) Result {
	filtered := &Array[T]{}
	for _, v := range s.items {
		if fn(v) {
			filtered.items = append(filtered.items, v)
		}
	}
	return Result{filtered, true}
}

// Each runs a function on every element.
//
//	agents := core.NewArray("codex", "hades")
//	agents.Each(func(name string) { core.Println(name) })
func (s *Array[T]) Each(fn func(T)) {
	for _, v := range s.items {
		fn(v)
	}
}

// Remove removes the first occurrence of a value.
//
//	agents := core.NewArray("codex", "hades", "homelab")
//	agents.Remove("hades")
func (s *Array[T]) Remove(val T) {
	for i, v := range s.items {
		if v == val {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return
		}
	}
}

// Deduplicate removes duplicate values, preserving order.
//
//	agents := core.NewArray("codex", "codex", "hades")
//	agents.Deduplicate()
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
//
//	agents := core.NewArray("codex", "hades")
//	count := agents.Len()
//	core.Println(count)
func (s *Array[T]) Len() int {
	return len(s.items)
}

// Clear removes all elements.
//
//	agents := core.NewArray("codex", "hades")
//	agents.Clear()
func (s *Array[T]) Clear() {
	s.items = nil
}

// AsSlice returns a copy of the underlying slice.
//
//	agents := core.NewArray("codex", "hades")
//	names := agents.AsSlice()
//	core.Println(core.Join(", ", names...))
func (s *Array[T]) AsSlice() []T {
	if s.items == nil {
		return nil
	}
	out := make([]T, len(s.items))
	copy(out, s.items)
	return out
}
