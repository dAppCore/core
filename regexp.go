// SPDX-License-Identifier: EUPL-1.2

// Compiled regular expression primitive for the Core framework.
// Wraps stdlib regexp with the Result-shape error pattern.
//
// Usage:
//
//	r := core.Regex(`\d+`)
//	if !r.OK { return r }
//	rx := r.Value.(*Regex)
//
//	rx.MatchString("foo123bar")     // true
//	rx.FindString("hello 42 world") // "42"
//	rx.FindAllString("a1 b2 c3", -1) // ["1","2","3"]
//	rx.ReplaceAllString("a1b2", "X") // "aXbX"
//	rx.Split("a,b,,c", -1)          // ["a","b","","c"]
package core

import "regexp"

// Regexp is a compiled regular expression. Construct with Regex(pattern).
// Named Regexp (matching stdlib) so the package-level constructor can be
// the bare verb Regex.
//
//	r := core.Regex(`agent-[0-9]+`)
//	if !r.OK { return r }
//	rx := r.Value.(*core.Regexp)
//	core.Println(rx.String())
type Regexp struct {
	inner *regexp.Regexp
}

// Regex compiles pattern. Returns Result wrapping *Regexp on success or
// the compile error if the pattern is invalid.
//
//	r := core.Regex(`\d+`)
//	if !r.OK { return r }
//	rx := r.Value.(*Regexp)
func Regex(pattern string) Result {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return Result{err, false}
	}
	return Result{&Regexp{inner: re}, true}
}

// MatchString reports whether s contains any match.
//
//	rx.MatchString("foo123bar")  // true
func (r *Regexp) MatchString(s string) bool {
	return r.inner.MatchString(s)
}

// FindString returns the leftmost match in s, or "" if none.
//
//	rx.FindString("hello 42 world")  // "42"
func (r *Regexp) FindString(s string) string {
	return r.inner.FindString(s)
}

// FindAllString returns successive matches. n controls max matches; n<0
// means all.
//
//	rx.FindAllString("a1 b2 c3", -1)  // ["1","2","3"]
func (r *Regexp) FindAllString(s string, n int) []string {
	return r.inner.FindAllString(s, n)
}

// FindStringSubmatch returns the leftmost match plus its submatches, or
// nil if no match.
//
//	rx := core.Regex(`(\w+)=(\d+)`).Value.(*core.Regexp)
//	rx.FindStringSubmatch("count=42")  // ["count=42", "count", "42"]
func (r *Regexp) FindStringSubmatch(s string) []string {
	return r.inner.FindStringSubmatch(s)
}

// ReplaceAllString returns a copy of s with all matches replaced by repl.
//
//	rx.ReplaceAllString("a1b2", "X")  // "aXbX"
func (r *Regexp) ReplaceAllString(s, repl string) string {
	return r.inner.ReplaceAllString(s, repl)
}

// Split splits s into substrings separated by the regex. n<0 means all.
//
//	rx := core.Regex(`,+`).Value.(*core.Regexp)
//	rx.Split("a,b,,c", -1)  // ["a","b","c"]
func (r *Regexp) Split(s string, n int) []string {
	return r.inner.Split(s, n)
}

// String returns the source pattern.
//
//	r := core.Regex(`agent-[0-9]+`)
//	if !r.OK { return r }
//	pattern := r.Value.(*core.Regexp).String()
//	core.Println(pattern)
func (r *Regexp) String() string {
	return r.inner.String()
}
