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

// Arg extracts a value from a variadic any slice at the given index.
// Returns nil if index is out of bounds.
//
//	val := core.Arg(args, 0)             // any
//	name := core.ArgString(args, 0)      // string
//	port := core.ArgInt(args, 1)         // int
//	debug := core.ArgBool(args, 2)       // bool
func Arg(args []any, index int) any {
	if index >= len(args) {
		return nil
	}
	return args[index]
}

// ArgString extracts a string from a variadic any slice at the given index.
//
//	name := core.ArgString(args, 0)
func ArgString(args []any, index int) string {
	v := Arg(args, index)
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

// ArgInt extracts an int from a variadic any slice at the given index.
//
//	port := core.ArgInt(args, 1)
func ArgInt(args []any, index int) int {
	v := Arg(args, index)
	if v == nil {
		return 0
	}
	i, _ := v.(int)
	return i
}

// ArgBool extracts a bool from a variadic any slice at the given index.
//
//	debug := core.ArgBool(args, 2)
func ArgBool(args []any, index int) bool {
	v := Arg(args, index)
	if v == nil {
		return false
	}
	b, _ := v.(bool)
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
