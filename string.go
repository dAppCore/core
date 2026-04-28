// SPDX-License-Identifier: EUPL-1.2

// String operations for the Core framework.
// Provides safe, predictable string helpers that downstream packages
// use directly — same pattern as Array[T] for slices.

package core

import (
	"strings"
	"unicode/utf8"
)

// HasPrefix returns true if s starts with prefix.
//
//	core.HasPrefix("--verbose", "--")  // true
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffix returns true if s ends with suffix.
//
//	core.HasSuffix("test.go", ".go")  // true
func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// TrimPrefix removes prefix from s.
//
//	core.TrimPrefix("--verbose", "--")  // "verbose"
func TrimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

// TrimSuffix removes suffix from s.
//
//	core.TrimSuffix("test.go", ".go")  // "test"
func TrimSuffix(s, suffix string) string {
	return strings.TrimSuffix(s, suffix)
}

// Contains returns true if s contains substr.
//
//	core.Contains("hello world", "world")  // true
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Split splits s by separator.
//
//	core.Split("a/b/c", "/")  // ["a", "b", "c"]
func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

// SplitN splits s by separator into at most n parts.
//
//	core.SplitN("key=value=extra", "=", 2)  // ["key", "value=extra"]
func SplitN(s, sep string, n int) []string {
	return strings.SplitN(s, sep, n)
}

// Join joins parts with a separator, building via Concat.
//
//	core.Join("/", "deploy", "to", "homelab")      // "deploy/to/homelab"
//	core.Join(".", "cmd", "deploy", "description")  // "cmd.deploy.description"
func Join(sep string, parts ...string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result = Concat(result, sep, p)
	}
	return result
}

// Replace replaces all occurrences of old with new in s.
//
//	core.Replace("deploy/to/homelab", "/", ".")  // "deploy.to.homelab"
func Replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// Lower returns s in lowercase.
//
//	core.Lower("HELLO")  // "hello"
func Lower(s string) string {
	return strings.ToLower(s)
}

// Upper returns s in uppercase.
//
//	core.Upper("hello")  // "HELLO"
func Upper(s string) string {
	return strings.ToUpper(s)
}

// Trim removes leading and trailing whitespace.
//
//	core.Trim("  hello  ")  // "hello"
func Trim(s string) string {
	return strings.TrimSpace(s)
}

// RuneCount returns the number of runes (unicode characters) in s.
//
//	core.RuneCount("hello")  // 5
//	core.RuneCount("🔥")     // 1
func RuneCount(s string) int {
	return utf8.RuneCountInString(s)
}

// NewBuilder returns a new strings.Builder.
//
//	b := core.NewBuilder()
//	b.WriteString("hello")
//	b.String() // "hello"
func NewBuilder() *strings.Builder {
	return &strings.Builder{}
}

// NewReader returns a strings.NewReader for the given string.
//
//	r := core.NewReader("hello world")
func NewReader(s string) *strings.Reader {
	return strings.NewReader(s)
}

// Concat joins variadic string parts into one string.
// Hook point for validation, sanitisation, and security checks.
//
//	core.Concat("cmd.", "deploy.to.homelab", ".description")
//	core.Concat("https://", host, "/api/v1")
func Concat(parts ...string) string {
	b := NewBuilder()
	for _, p := range parts {
		b.WriteString(p)
	}
	return b.String()
}
