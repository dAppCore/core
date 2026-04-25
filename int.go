// SPDX-License-Identifier: EUPL-1.2

// Integer conversion helpers for the Core framework.
// Wraps strconv so consumers don't import it directly.
package core

import "strconv"

// Atoi converts a decimal string to an int.
//
//	r := core.Atoi("42")
//	if r.OK { n := r.Value.(int) }
func Atoi(s string) Result {
	i, err := strconv.Atoi(s)
	if err != nil {
		return Result{err, false}
	}
	return Result{i, true}
}

// Itoa converts an int to a decimal string.
//
//	s := core.Itoa(42)
func Itoa(i int) string {
	return strconv.Itoa(i)
}

// FormatInt converts an int64 to a string in the given base.
//
//	s := core.FormatInt(255, 16) // "ff"
func FormatInt(i int64, base int) string {
	return strconv.FormatInt(i, base)
}

// ParseInt converts a string in the given base and bit size to an int64.
//
//	r := core.ParseInt("ff", 16, 64)
//	if r.OK { n := r.Value.(int64) }
func ParseInt(s string, base int, bitSize int) Result {
	i, err := strconv.ParseInt(s, base, bitSize)
	if err != nil {
		return Result{err, false}
	}
	return Result{i, true}
}
