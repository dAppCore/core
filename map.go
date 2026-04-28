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

// MapFilter returns a new map containing only entries for which pred
// returns true. A nil result is returned for an empty input.
//
//	enabled := core.MapFilter(features, func(name string, on bool) bool { return on })
func MapFilter[K comparable, V any](m map[K]V, pred func(K, V) bool) map[K]V {
	if len(m) == 0 {
		return nil
	}
	out := make(map[K]V, len(m))
	for k, v := range m {
		if pred(k, v) {
			out[k] = v
		}
	}
	return out
}

// MapMerge returns a new map containing all entries from a and b.
// When a key appears in both, b's value wins.
//
//	combined := core.MapMerge(defaults, overrides)
func MapMerge[K comparable, V any](a, b map[K]V) map[K]V {
	out := make(map[K]V, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

// MapHasKey reports whether k is present in m.
//
//	if core.MapHasKey(config, "debug") { ... }
func MapHasKey[K comparable, V any](m map[K]V, k K) bool {
	_, ok := m[k]
	return ok
}
