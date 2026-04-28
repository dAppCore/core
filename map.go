// SPDX-License-Identifier: EUPL-1.2

// Map helpers for the Core framework.

package core

import "maps"

// MapKeys returns the keys from m.
// Map iteration order is not stable; sort the result when order matters.
//
//	keys := core.MapKeys(map[string]int{"a": 1, "b": 2})
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

// MapValues returns the values from m.
// Map iteration order is not stable; sort the result when order matters.
//
//	values := core.MapValues(map[string]int{"a": 1, "b": 2})
func MapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, value := range m {
		values = append(values, value)
	}
	return values
}

// MapClone returns a shallow clone of m.
// A nil map remains nil.
//
//	copy := core.MapClone(map[string]int{"a": 1})
func MapClone[K comparable, V any](m map[K]V) map[K]V {
	return maps.Clone(m)
}
