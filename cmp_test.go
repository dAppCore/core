package core_test

import (
	. "dappco.re/go"
)

func TestCmp_Compare_Good(t *T) {
	AssertEqual(t, -1, Compare(1, 2))
	AssertEqual(t, 1, Compare(2, 1))
}

func TestCmp_Compare_Bad(t *T) {
	AssertEqual(t, 0, Compare(7, 7))
}

func TestCmp_Compare_Ugly(t *T) {
	AssertEqual(t, -1, Compare("alpha", "beta"))
	AssertEqual(t, 1, Compare("beta", "alpha"))
}
