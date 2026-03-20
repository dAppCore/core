// SPDX-License-Identifier: EUPL-1.2

// Utility functions for the Core framework.
// Built on core string.go primitives.

package core

import (
	"fmt"
	"io"
	"os"
)

// Print writes a formatted line to a writer, defaulting to os.Stdout.
//
//	core.Print(nil, "hello %s", "world")     // → stdout
//	core.Print(w, "port: %d", 8080)          // → w
func Print(w io.Writer, format string, args ...any) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, format+"\n", args...)
}

// JoinPath joins string segments into a path with "/" separator.
//
//	core.JoinPath("deploy", "to", "homelab")  // → "deploy/to/homelab"
func JoinPath(segments ...string) string {
	return Join("/", segments...)
}

// IsFlag returns true if the argument starts with a dash.
//
//	core.IsFlag("--verbose")  // true
//	core.IsFlag("-v")         // true
//	core.IsFlag("deploy")    // false
func IsFlag(arg string) bool {
	return HasPrefix(arg, "-")
}

// Arg extracts a value from variadic args at the given index.
// Type-checks and delegates to the appropriate typed extractor.
// Returns Result — OK is false if index is out of bounds.
//
//	r := core.Arg(0, args...)
//	if r.OK { path = r.Value.(string) }
func Arg(index int, args ...any) Result {
	if index >= len(args) {
		return Result{}
	}
	v := args[index]
	switch v.(type) {
	case string:
		return Result{Value: ArgString(index, args...), OK: true}
	case int:
		return Result{Value: ArgInt(index, args...), OK: true}
	case bool:
		return Result{Value: ArgBool(index, args...), OK: true}
	default:
		return Result{Value: v, OK: true}
	}
}

// ArgString extracts a string at the given index.
//
//	name := core.ArgString(0, args...)
func ArgString(index int, args ...any) string {
	if index >= len(args) {
		return ""
	}
	s, ok := args[index].(string)
	if !ok {
		return ""
	}
	return s
}

// ArgInt extracts an int at the given index.
//
//	port := core.ArgInt(1, args...)
func ArgInt(index int, args ...any) int {
	if index >= len(args) {
		return 0
	}
	i, ok := args[index].(int)
	if !ok {
		return 0
	}
	return i
}

// ArgBool extracts a bool at the given index.
//
//	debug := core.ArgBool(2, args...)
func ArgBool(index int, args ...any) bool {
	if index >= len(args) {
		return false
	}
	b, ok := args[index].(bool)
	if !ok {
		return false
	}
	return b
}

// FilterArgs removes empty strings and Go test runner flags from an argument list.
//
//	clean := core.FilterArgs(os.Args[1:])
func FilterArgs(args []string) []string {
	var clean []string
	for _, a := range args {
		if a == "" || HasPrefix(a, "-test.") {
			continue
		}
		clean = append(clean, a)
	}
	return clean
}

// ParseFlag parses a single flag argument into key, value, and validity.
// Single dash (-) requires exactly 1 character (letter, emoji, unicode).
// Double dash (--) requires 2+ characters.
//
//	"-v"           → "v", "", true
//	"-🔥"          → "🔥", "", true
//	"--verbose"    → "verbose", "", true
//	"--port=8080"  → "port", "8080", true
//	"-verbose"     → "", "", false  (single dash, 2+ chars)
//	"--v"          → "", "", false  (double dash, 1 char)
//	"hello"        → "", "", false  (not a flag)
func ParseFlag(arg string) (key, value string, valid bool) {
	if HasPrefix(arg, "--") {
		rest := TrimPrefix(arg, "--")
		parts := SplitN(rest, "=", 2)
		name := parts[0]
		if RuneCount(name) < 2 {
			return "", "", false
		}
		if len(parts) == 2 {
			return name, parts[1], true
		}
		return name, "", true
	}

	if HasPrefix(arg, "-") {
		rest := TrimPrefix(arg, "-")
		parts := SplitN(rest, "=", 2)
		name := parts[0]
		if RuneCount(name) != 1 {
			return "", "", false
		}
		if len(parts) == 2 {
			return name, parts[1], true
		}
		return name, "", true
	}

	return "", "", false
}
