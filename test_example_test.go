package core_test

import . "dappco.re/go"

// ExampleAnError uses the sentinel error through `AnError` for AX-native tests. Passing
// assertions are silent while failures stay one-line and AI-readable.
func ExampleAnError() {
	Println(AnError.Error())
	// Output: core test sentinel error
}

// ExampleT documents the testing alias through `T` for AX-native tests. Passing assertions
// are silent while failures stay one-line and AI-readable.
func ExampleT() {
	var t *T
	_ = t
}

// ExampleTB documents the testing interface alias through `TB` for AX-native tests.
// Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleTB() {
	var t *T
	var tb TB = t
	_ = tb
}

// ExampleB documents the benchmark alias through `B` for AX-native tests. Passing
// assertions are silent while failures stay one-line and AI-readable.
func ExampleB() {
	var b *B
	_ = b
}

// ExampleAssertEqual asserts equal values through `AssertEqual` for AX-native tests.
// Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertEqual() {
	var t *T
	AssertEqual(t, "expected", "expected")
}

// ExampleAssertNotEqual asserts different values through `AssertNotEqual` for AX-native
// tests. Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertNotEqual() {
	var t *T
	AssertNotEqual(t, "old", "new")
}

// ExampleAssertTrue asserts a true condition through `AssertTrue` for AX-native tests.
// Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertTrue() {
	var t *T
	AssertTrue(t, true)
}

// ExampleAssertFalse asserts a false condition through `AssertFalse` for AX-native tests.
// Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertFalse() {
	var t *T
	AssertFalse(t, false)
}

// ExampleAssertNil asserts nil through `AssertNil` for AX-native tests. Passing assertions
// are silent while failures stay one-line and AI-readable.
func ExampleAssertNil() {
	var t *T
	AssertNil(t, nil)
}

// ExampleAssertNotNil asserts a non-nil value through `AssertNotNil` for AX-native tests.
// Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertNotNil() {
	var t *T
	AssertNotNil(t, "value")
}

// ExampleAssertNoError asserts no error through `AssertNoError` for AX-native tests.
// Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertNoError() {
	var t *T
	AssertNoError(t, nil)
}

// ExampleAssertError asserts an error through `AssertError` for AX-native tests. Passing
// assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertError() {
	var t *T
	AssertError(t, AnError, "sentinel")
}

// ExampleAssertContains asserts membership through `AssertContains` for AX-native tests.
// Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertContains() {
	var t *T
	AssertContains(t, []string{"alpha", "bravo"}, "bravo")
}

// ExampleAssertNotContains asserts absence through `AssertNotContains` for AX-native
// tests. Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertNotContains() {
	var t *T
	AssertNotContains(t, []string{"alpha", "bravo"}, "charlie")
}

// ExampleAssertLen asserts length through `AssertLen` for AX-native tests. Passing
// assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertLen() {
	var t *T
	AssertLen(t, []string{"alpha", "bravo"}, 2)
}

// ExampleAssertEmpty asserts emptiness through `AssertEmpty` for AX-native tests. Passing
// assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertEmpty() {
	var t *T
	AssertEmpty(t, []string{})
}

// ExampleAssertNotEmpty asserts non-emptiness through `AssertNotEmpty` for AX-native
// tests. Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertNotEmpty() {
	var t *T
	AssertNotEmpty(t, []string{"alpha"})
}

// ExampleAssertGreater asserts greater-than order through `AssertGreater` for AX-native
// tests. Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertGreater() {
	var t *T
	AssertGreater(t, 3, 2)
}

// ExampleAssertGreaterOrEqual asserts greater-or-equal order through
// `AssertGreaterOrEqual` for AX-native tests. Passing assertions are silent while failures
// stay one-line and AI-readable.
func ExampleAssertGreaterOrEqual() {
	var t *T
	AssertGreaterOrEqual(t, 3, 3)
}

// ExampleAssertLess asserts less-than order through `AssertLess` for AX-native tests.
// Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertLess() {
	var t *T
	AssertLess(t, 2, 3)
}

// ExampleAssertLessOrEqual asserts less-or-equal order through `AssertLessOrEqual` for
// AX-native tests. Passing assertions are silent while failures stay one-line and
// AI-readable.
func ExampleAssertLessOrEqual() {
	var t *T
	AssertLessOrEqual(t, 3, 3)
}

// ExampleAssertPanics asserts panic behaviour through `AssertPanics` for AX-native tests.
// Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertPanics() {
	var t *T
	AssertPanics(t, func() { panic("boom") })
}

// ExampleAssertNotPanics asserts non-panicking behaviour through `AssertNotPanics` for
// AX-native tests. Passing assertions are silent while failures stay one-line and
// AI-readable.
func ExampleAssertNotPanics() {
	var t *T
	AssertNotPanics(t, func() { /* no-op closure demonstrates non-panicking call */ })
}

// ExampleAssertPanicsWithError asserts panic error text through `AssertPanicsWithError`
// for AX-native tests. Passing assertions are silent while failures stay one-line and
// AI-readable.
func ExampleAssertPanicsWithError() {
	var t *T
	AssertPanicsWithError(t, "sentinel", func() { panic(AnError) })
}

// ExampleAssertErrorIs asserts wrapped error identity through `AssertErrorIs` for
// AX-native tests. Passing assertions are silent while failures stay one-line and
// AI-readable.
func ExampleAssertErrorIs() {
	var t *T
	AssertErrorIs(t, Wrap(AnError, "example", "wrapped"), AnError)
}

// ExampleAssertInDelta asserts approximate numeric equality through `AssertInDelta` for
// AX-native tests. Passing assertions are silent while failures stay one-line and
// AI-readable.
func ExampleAssertInDelta() {
	var t *T
	AssertInDelta(t, 1.0, 1.01, 0.02)
}

// ExampleAssertSame asserts pointer identity through `AssertSame` for AX-native tests.
// Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleAssertSame() {
	var t *T
	a := &struct{}{}
	AssertSame(t, a, a)
}

// ExampleRequireNoError requires no error through `RequireNoError` for AX-native tests.
// Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleRequireNoError() {
	var t *T
	RequireNoError(t, nil)
}

// ExampleRequireTrue requires a true condition through `RequireTrue` for AX-native tests.
// Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleRequireTrue() {
	var t *T
	RequireTrue(t, true)
}

// ExampleRequireNotEmpty requires non-empty input through `RequireNotEmpty` for AX-native
// tests. Passing assertions are silent while failures stay one-line and AI-readable.
func ExampleRequireNotEmpty() {
	var t *T
	RequireNotEmpty(t, []string{"alpha"})
}

// ExampleAssertElementsMatch asserts unordered list equality through `AssertElementsMatch`
// for AX-native tests. Passing assertions are silent while failures stay one-line and
// AI-readable.
func ExampleAssertElementsMatch() {
	var t *T
	AssertElementsMatch(t, []string{"alpha", "bravo"}, []string{"bravo", "alpha"})
}
