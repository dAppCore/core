// SPDX-License-Identifier: EUPL-1.2

// Iterator primitives — re-exports of Go's iter package as Core types,
// so consumers never need to write `import "iter"`. Sequences yielded
// by core helpers (Fs.WalkSeq, Slice generators, etc.) all use these
// aliases.
//
//	for path := range c.Fs().WalkSeq(root) {
//	    Println(path)
//	}
package core

import "iter"

// Seq is a single-value iterator — a function that pushes elements one
// at a time to a yield callback until it returns false. Alias of iter.Seq.
//
//	var nums core.Seq[int] = func(yield func(int) bool) {
//	    for i := range 5 { if !yield(i) { return } }
//	}
//	for n := range nums { core.Println(n) }
type Seq[V any] = iter.Seq[V]

// Seq2 is a two-value iterator — yielding (key, value) or (index, item)
// pairs. Alias of iter.Seq2. Returned by walkers and key/value scanners.
//
//	var entries core.Seq2[string, int] = func(yield func(string, int) bool) {
//	    if !yield("a", 1) { return }
//	    if !yield("b", 2) { return }
//	}
//	for k, v := range entries { core.Println(k, v) }
type Seq2[K, V any] = iter.Seq2[K, V]

// Pull converts a push-style Seq into a pull-style (next, stop) pair.
// next() returns (value, ok); stop() releases iterator resources.
//
//	next, stop := core.Pull(seq)
//	defer stop()
//	for {
//	    v, ok := next()
//	    if !ok { break }
//	    core.Println(v)
//	}
func Pull[V any](seq Seq[V]) (next func() (V, bool), stop func()) {
	return iter.Pull(seq)
}

// Pull2 converts a push-style Seq2 into a pull-style (next, stop) pair.
// next() returns (key, value, ok); stop() releases iterator resources.
//
//	next, stop := core.Pull2(entries)
//	defer stop()
//	for {
//	    k, v, ok := next()
//	    if !ok { break }
//	    core.Println(k, v)
//	}
func Pull2[K, V any](seq Seq2[K, V]) (next func() (K, V, bool), stop func()) {
	return iter.Pull2(seq)
}
