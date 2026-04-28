package core_test

import . "dappco.re/go"

func ExampleAnError() {
	Println(AnError.Error())
	// Output: core test sentinel error
}

func ExampleT() {
	var t *T
	_ = t
}

func ExampleTB() {
	var t *T
	var tb TB = t
	_ = tb
}

func ExampleB() {
	var b *B
	_ = b
}

func ExampleAssertEqual() {
	var t *T
	AssertEqual(t, "expected", "expected")
}

func ExampleAssertNotEqual() {
	var t *T
	AssertNotEqual(t, "old", "new")
}

func ExampleAssertTrue() {
	var t *T
	AssertTrue(t, true)
}

func ExampleAssertFalse() {
	var t *T
	AssertFalse(t, false)
}

func ExampleAssertNil() {
	var t *T
	AssertNil(t, nil)
}

func ExampleAssertNotNil() {
	var t *T
	AssertNotNil(t, "value")
}

func ExampleAssertNoError() {
	var t *T
	AssertNoError(t, nil)
}

func ExampleAssertError() {
	var t *T
	AssertError(t, AnError, "sentinel")
}

func ExampleAssertContains() {
	var t *T
	AssertContains(t, []string{"alpha", "bravo"}, "bravo")
}

func ExampleAssertNotContains() {
	var t *T
	AssertNotContains(t, []string{"alpha", "bravo"}, "charlie")
}

func ExampleAssertLen() {
	var t *T
	AssertLen(t, []string{"alpha", "bravo"}, 2)
}

func ExampleAssertEmpty() {
	var t *T
	AssertEmpty(t, []string{})
}

func ExampleAssertNotEmpty() {
	var t *T
	AssertNotEmpty(t, []string{"alpha"})
}

func ExampleAssertGreater() {
	var t *T
	AssertGreater(t, 3, 2)
}

func ExampleAssertGreaterOrEqual() {
	var t *T
	AssertGreaterOrEqual(t, 3, 3)
}

func ExampleAssertLess() {
	var t *T
	AssertLess(t, 2, 3)
}

func ExampleAssertLessOrEqual() {
	var t *T
	AssertLessOrEqual(t, 3, 3)
}

func ExampleAssertPanics() {
	var t *T
	AssertPanics(t, func() { panic("boom") })
}

func ExampleAssertNotPanics() {
	var t *T
	AssertNotPanics(t, func() {})
}

func ExampleAssertPanicsWithError() {
	var t *T
	AssertPanicsWithError(t, "sentinel", func() { panic(AnError) })
}

func ExampleAssertErrorIs() {
	var t *T
	AssertErrorIs(t, Wrap(AnError, "example", "wrapped"), AnError)
}

func ExampleAssertInDelta() {
	var t *T
	AssertInDelta(t, 1.0, 1.01, 0.02)
}

func ExampleAssertSame() {
	var t *T
	a := &struct{}{}
	AssertSame(t, a, a)
}

func ExampleRequireNoError() {
	var t *T
	RequireNoError(t, nil)
}

func ExampleRequireTrue() {
	var t *T
	RequireTrue(t, true)
}

func ExampleRequireNotEmpty() {
	var t *T
	RequireNotEmpty(t, []string{"alpha"})
}

func ExampleAssertElementsMatch() {
	var t *T
	AssertElementsMatch(t, []string{"alpha", "bravo"}, []string{"bravo", "alpha"})
}
