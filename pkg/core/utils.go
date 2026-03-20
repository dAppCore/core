// SPDX-License-Identifier: EUPL-1.2

// Utility functions for the Core framework.

package core

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// Printl writes a formatted line to a writer, defaulting to os.Stdout.
//
//	core.Printl(nil, "hello %s", "world")     // → stdout
//	core.Printl(w, "port: %d", 8080)          // → w
func Printl(w io.Writer, format string, args ...any) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, format+"\n", args...)
}

// FilterArgs removes empty strings and Go test runner flags from an argument list.
//
//	clean := core.FilterArgs(os.Args[1:])
func FilterArgs(args []string) []string {
	var clean []string
	for _, a := range args {
		if a == "" || strings.HasPrefix(a, "-test.") {
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
	if strings.HasPrefix(arg, "--") {
		// Long flag: must be 2+ chars
		rest := strings.TrimPrefix(arg, "--")
		parts := strings.SplitN(rest, "=", 2)
		name := parts[0]
		if utf8.RuneCountInString(name) < 2 {
			return "", "", false
		}
		if len(parts) == 2 {
			return name, parts[1], true
		}
		return name, "", true
	}

	if strings.HasPrefix(arg, "-") {
		// Short flag: must be exactly 1 char (rune)
		rest := strings.TrimPrefix(arg, "-")
		parts := strings.SplitN(rest, "=", 2)
		name := parts[0]
		if utf8.RuneCountInString(name) != 1 {
			return "", "", false
		}
		if len(parts) == 2 {
			return name, parts[1], true
		}
		return name, "", true
	}

	return "", "", false
}
