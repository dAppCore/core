package core_test

import (
	. "dappco.re/go"
)

func TestIter_Pull_Good(t *T) {
	seq := func(yield func(string) bool) {
		yield("codex")
		yield("hades")
	}
	next, stop := Pull(seq)
	defer stop()

	first, ok := next()
	AssertTrue(t, ok)
	AssertEqual(t, "codex", first)
	second, ok := next()
	AssertTrue(t, ok)
	AssertEqual(t, "hades", second)
	_, ok = next()
	AssertFalse(t, ok)
}

func TestIter_Pull_Bad(t *T) {
	seq := func(yield func(string) bool) { /* empty sequence yields nothing */ }
	next, stop := Pull(seq)
	defer stop()

	value, ok := next()

	AssertFalse(t, ok)
	AssertEqual(t, "", value)
}

func TestIter_Pull_Ugly(t *T) {
	seq := func(yield func(int) bool) {
		for i := 0; i < 3; i++ {
			if !yield(i) {
				return
			}
		}
	}
	next, stop := Pull(seq)

	value, ok := next()
	stop()
	after, afterOK := next()

	AssertTrue(t, ok)
	AssertEqual(t, 0, value)
	AssertFalse(t, afterOK)
	AssertEqual(t, 0, after)
}

func TestIter_Pull2_Good(t *T) {
	seq := func(yield func(string, int) bool) {
		yield("codex", 1)
		yield("hades", 2)
	}
	next, stop := Pull2(seq)
	defer stop()

	key, value, ok := next()
	AssertTrue(t, ok)
	AssertEqual(t, "codex", key)
	AssertEqual(t, 1, value)
	key, value, ok = next()
	AssertTrue(t, ok)
	AssertEqual(t, "hades", key)
	AssertEqual(t, 2, value)
}

func TestIter_Pull2_Bad(t *T) {
	seq := func(yield func(string, int) bool) { /* empty sequence yields nothing */ }
	next, stop := Pull2(seq)
	defer stop()

	key, value, ok := next()

	AssertFalse(t, ok)
	AssertEqual(t, "", key)
	AssertEqual(t, 0, value)
}

func TestIter_Pull2_Ugly(t *T) {
	seq := func(yield func(string, int) bool) {
		yield("codex", 1)
		yield("hades", 2)
	}
	next, stop := Pull2(seq)

	key, value, ok := next()
	stop()
	afterKey, afterValue, afterOK := next()

	AssertTrue(t, ok)
	AssertEqual(t, "codex", key)
	AssertEqual(t, 1, value)
	AssertFalse(t, afterOK)
	AssertEqual(t, "", afterKey)
	AssertEqual(t, 0, afterValue)
}
