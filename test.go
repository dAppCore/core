// SPDX-License-Identifier: EUPL-1.2

// AX-shaped test assertions for the Core framework.
//
// Replaces testify usage in core/go and downstream consumer packages.
// Thin wrappers around testing.T (which stays as the runner — Go's test
// framework is the keep-exemption #1104) that emit terse, value-added
// failure messages suitable for AI-driven test triage.
//
// On pass:  silent (Go test default).
// On fail:  one line, file:line prepended by go test, then a key=value
//
//	shape an AI agent can grep for assertion / want / got / msg.
//
// Format:
//
//	fs_test.go:144: AssertEqual want="hello" got="world"
//	fs_test.go:147: AssertNoError got=open foo: no such file or directory
//
// Usage:
//
//	core.AssertEqual(t, "expected", actual)
//	core.AssertTrue(t, result.OK)
//	core.AssertNoError(t, err)
//	core.AssertContains(t, []string{"a", "b"}, "a")
//
// All assertions accept testing.TB so they work in both Test* and
// Benchmark* functions. Optional trailing string args are joined into
// a context suffix on the failure line.
package core

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

// AnError is a sentinel error for tests that need a non-nil error
// without caring about its content. Mirrors testify's assert.AnError.
//
//	core.AssertError(t, somethingThatFails(), core.AnError.Error())
var AnError = errors.New("core test sentinel error")

// T is the canonical Go test handle, exported as core.T so test files
// don't need a separate `import "testing"` line. Go's test runner
// accepts *core.T in TestXxx signatures because the alias is
// type-identical to *testing.T.
//
//	func TestSomething_Good(t *core.T) {
//	    core.AssertEqual(t, expected, actual)
//	}
type T = testing.T

// TB is the testing-handle interface (T + B), exported as core.TB so
// helpers can accept either Test or Benchmark contexts without
// importing testing.
//
//	func helper(t core.TB, ...) { t.Helper(); ... }
type TB = testing.TB

// B is the canonical Go benchmark handle, exported as core.B.
//
//	func BenchmarkSomething(b *core.B) { ... }
type B = testing.B

// AssertEqual fails the test if want and got are not deeply equal.
//
//	core.AssertEqual(t, "expected", result.Value)
//	core.AssertEqual(t, 42, result.Count)
func AssertEqual(t testing.TB, want, got any, msg ...string) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("AssertEqual want=%#v got=%#v%s", want, got, assertMsg(msg))
	}
}

// AssertNotEqual fails the test if want and got are deeply equal.
//
//	core.AssertNotEqual(t, oldValue, newValue)
func AssertNotEqual(t testing.TB, want, got any, msg ...string) {
	t.Helper()
	if reflect.DeepEqual(want, got) {
		t.Errorf("AssertNotEqual want!=%#v got=%#v%s", want, got, assertMsg(msg))
	}
}

// AssertTrue fails the test if condition is false.
//
//	core.AssertTrue(t, result.OK)
//	core.AssertTrue(t, len(items) > 0, "items must not be empty")
func AssertTrue(t testing.TB, condition bool, msg ...string) {
	t.Helper()
	if !condition {
		t.Errorf("AssertTrue want=true got=false%s", assertMsg(msg))
	}
}

// AssertFalse fails the test if condition is true.
//
//	core.AssertFalse(t, c.Fs().Exists(deletedPath))
func AssertFalse(t testing.TB, condition bool, msg ...string) {
	t.Helper()
	if condition {
		t.Errorf("AssertFalse want=false got=true%s", assertMsg(msg))
	}
}

// AssertNil fails the test if v is non-nil. Handles typed-nil interface
// values (e.g. a (*Foo)(nil) inside an any).
//
//	core.AssertNil(t, returnedPointer)
func AssertNil(t testing.TB, v any, msg ...string) {
	t.Helper()
	if !assertIsNil(v) {
		t.Errorf("AssertNil want=nil got=%#v%s", v, assertMsg(msg))
	}
}

// AssertNotNil fails the test if v is nil.
//
//	core.AssertNotNil(t, c.Fs())
func AssertNotNil(t testing.TB, v any, msg ...string) {
	t.Helper()
	if assertIsNil(v) {
		t.Errorf("AssertNotNil want=non-nil got=nil%s", assertMsg(msg))
	}
}

// AssertNoError fails the test if err is non-nil.
//
//	core.AssertNoError(t, c.Fs().Write(p, "data").Error())
func AssertNoError(t testing.TB, err error, msg ...string) {
	t.Helper()
	if err != nil {
		t.Errorf("AssertNoError got=%v%s", err, assertMsg(msg))
	}
}

// AssertError fails the test if err is nil. Optional substring matches
// against err.Error() for tighter assertions.
//
//	core.AssertError(t, parseFails())
//	core.AssertError(t, parseFails(), "invalid syntax")
func AssertError(t testing.TB, err error, msg ...string) {
	t.Helper()
	if err == nil {
		t.Errorf("AssertError want=non-nil got=nil%s", assertMsg(msg))
		return
	}
	for _, want := range msg {
		if !Contains(err.Error(), want) {
			t.Errorf("AssertError want-substring=%q got=%q", want, err.Error())
			return
		}
	}
}

// AssertContains fails the test if needle is not present in haystack.
// Supports strings (substring match), slices/arrays (deep-equal element
// membership), and maps (key membership).
//
//	core.AssertContains(t, "hello world", "world")
//	core.AssertContains(t, []string{"a", "b"}, "a")
//	core.AssertContains(t, map[string]int{"x": 1}, "x")
func AssertContains(t testing.TB, haystack, needle any, msg ...string) {
	t.Helper()
	if !assertContains(haystack, needle) {
		t.Errorf("AssertContains haystack=%#v needle=%#v%s", haystack, needle, assertMsg(msg))
	}
}

// AssertNotContains fails the test if needle IS present in haystack.
func AssertNotContains(t testing.TB, haystack, needle any, msg ...string) {
	t.Helper()
	if assertContains(haystack, needle) {
		t.Errorf("AssertNotContains haystack=%#v needle=%#v (found)%s", haystack, needle, assertMsg(msg))
	}
}

// AssertLen fails the test if the length of v does not equal want.
// Works on strings, slices, arrays, maps, and channels.
//
//	core.AssertLen(t, items, 3)
func AssertLen(t testing.TB, v any, want int, msg ...string) {
	t.Helper()
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		if rv.Len() != want {
			t.Errorf("AssertLen want=%d got=%d%s", want, rv.Len(), assertMsg(msg))
		}
	default:
		t.Errorf("AssertLen unsupported kind=%s%s", rv.Kind(), assertMsg(msg))
	}
}

// AssertEmpty fails the test if v is not empty. Treats zero-length
// strings/slices/arrays/maps/channels and nil as empty.
//
//	core.AssertEmpty(t, results)
func AssertEmpty(t testing.TB, v any, msg ...string) {
	t.Helper()
	if !assertIsEmpty(v) {
		t.Errorf("AssertEmpty want=empty got=%#v%s", v, assertMsg(msg))
	}
}

// AssertNotEmpty fails the test if v is empty (see AssertEmpty).
//
//	core.AssertNotEmpty(t, response.Body)
func AssertNotEmpty(t testing.TB, v any, msg ...string) {
	t.Helper()
	if assertIsEmpty(v) {
		t.Errorf("AssertNotEmpty want=non-empty got=%#v%s", v, assertMsg(msg))
	}
}

// AssertGreater fails the test if got is not strictly greater than want.
// Works on numeric kinds (int, uint, float) and strings.
//
//	core.AssertGreater(t, count, 0)
func AssertGreater(t testing.TB, got, want any, msg ...string) {
	t.Helper()
	cmp, ok := assertCompare(got, want)
	if !ok {
		t.Errorf("AssertGreater incomparable got=%#v want=%#v%s", got, want, assertMsg(msg))
		return
	}
	if cmp <= 0 {
		t.Errorf("AssertGreater got=%#v want>%#v%s", got, want, assertMsg(msg))
	}
}

// AssertGreaterOrEqual fails the test if got is less than want.
//
//	core.AssertGreaterOrEqual(t, elapsed, minDuration)
func AssertGreaterOrEqual(t testing.TB, got, want any, msg ...string) {
	t.Helper()
	cmp, ok := assertCompare(got, want)
	if !ok {
		t.Errorf("AssertGreaterOrEqual incomparable got=%#v want=%#v%s", got, want, assertMsg(msg))
		return
	}
	if cmp < 0 {
		t.Errorf("AssertGreaterOrEqual got=%#v want>=%#v%s", got, want, assertMsg(msg))
	}
}

// AssertLess fails the test if got is not strictly less than want.
//
//	core.AssertLess(t, errorCount, limit)
func AssertLess(t testing.TB, got, want any, msg ...string) {
	t.Helper()
	cmp, ok := assertCompare(got, want)
	if !ok {
		t.Errorf("AssertLess incomparable got=%#v want=%#v%s", got, want, assertMsg(msg))
		return
	}
	if cmp >= 0 {
		t.Errorf("AssertLess got=%#v want<%#v%s", got, want, assertMsg(msg))
	}
}

// AssertLessOrEqual fails the test if got is greater than want.
func AssertLessOrEqual(t testing.TB, got, want any, msg ...string) {
	t.Helper()
	cmp, ok := assertCompare(got, want)
	if !ok {
		t.Errorf("AssertLessOrEqual incomparable got=%#v want=%#v%s", got, want, assertMsg(msg))
		return
	}
	if cmp > 0 {
		t.Errorf("AssertLessOrEqual got=%#v want<=%#v%s", got, want, assertMsg(msg))
	}
}

// AssertPanics fails the test if calling fn does not panic.
//
//	core.AssertPanics(t, func() { mustParse("garbage") })
func AssertPanics(t testing.TB, fn func(), msg ...string) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("AssertPanics want=panic got=normal-return%s", assertMsg(msg))
		}
	}()
	fn()
}

// AssertNotPanics fails the test if calling fn panics.
//
//	core.AssertNotPanics(t, func() { safeDivide(10, 2) })
func AssertNotPanics(t testing.TB, fn func(), msg ...string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AssertNotPanics want=normal-return got=panic(%v)%s", r, assertMsg(msg))
		}
	}()
	fn()
}

// AssertPanicsWithError fails the test if fn does not panic, or panics
// with a value whose error string does not contain wantSubstr. Argument
// order matches testify's PanicsWithError(t, errString, fn).
//
//	core.AssertPanicsWithError(t, "empty input", func() { mustParse("") })
func AssertPanicsWithError(t testing.TB, wantSubstr string, fn func(), msg ...string) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("AssertPanicsWithError want=panic got=normal-return%s", assertMsg(msg))
			return
		}
		var got string
		if err, ok := r.(error); ok {
			got = err.Error()
		} else {
			got = Sprintf("%v", r)
		}
		if !Contains(got, wantSubstr) {
			t.Errorf("AssertPanicsWithError want-substring=%q got=%q%s", wantSubstr, got, assertMsg(msg))
		}
	}()
	fn()
}

// AssertErrorIs fails the test if err does not wrap target (errors.Is).
//
//	core.AssertErrorIs(t, err, fs.ErrNotExist)
func AssertErrorIs(t testing.TB, err, target error, msg ...string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("AssertErrorIs want-target=%v got=%v%s", target, err, assertMsg(msg))
	}
}

// AssertInDelta fails the test if |got - want| > delta. Use for float
// comparisons where exact equality is not appropriate.
//
//	core.AssertInDelta(t, expected, actual, 0.0001)
func AssertInDelta(t testing.TB, want, got, delta float64, msg ...string) {
	t.Helper()
	if math.IsNaN(want) || math.IsNaN(got) {
		t.Errorf("AssertInDelta NaN involved want=%v got=%v%s", want, got, assertMsg(msg))
		return
	}
	diff := math.Abs(want - got)
	if diff > delta {
		t.Errorf("AssertInDelta want=%v got=%v delta=%v actual-diff=%v%s", want, got, delta, diff, assertMsg(msg))
	}
}

// AssertSame fails the test if want and got are not the same pointer.
//
//	core.AssertSame(t, c.Fs(), c.Fs())  // singleton check
func AssertSame(t testing.TB, want, got any, msg ...string) {
	t.Helper()
	wv := reflect.ValueOf(want)
	gv := reflect.ValueOf(got)
	if wv.Kind() != reflect.Ptr || gv.Kind() != reflect.Ptr {
		t.Errorf("AssertSame both args must be pointers, got want=%s got=%s%s", wv.Kind(), gv.Kind(), assertMsg(msg))
		return
	}
	if wv.Pointer() != gv.Pointer() {
		t.Errorf("AssertSame want=%p got=%p%s", want, got, assertMsg(msg))
	}
}

// RequireNoError fails the test AND stops it if err is non-nil. Use when
// the rest of the test depends on the operation succeeding.
//
//	core.RequireNoError(t, c.Fs().Write(p, "data").Error())
func RequireNoError(t testing.TB, err error, msg ...string) {
	t.Helper()
	if err != nil {
		t.Fatalf("RequireNoError got=%v%s", err, assertMsg(msg))
	}
}

// RequireTrue fails the test AND stops it if condition is false. Use to
// guard test invariants where continuing past a false condition is
// meaningless.
//
//	core.RequireTrue(t, scanner.Scan(), "scan precondition")
func RequireTrue(t testing.TB, condition bool, msg ...string) {
	t.Helper()
	if !condition {
		t.Fatalf("RequireTrue want=true got=false%s", assertMsg(msg))
	}
}

// RequireNotEmpty fails the test AND stops it if v is empty. Use to
// guard test invariants where downstream assertions presume non-empty
// fixture data.
//
//	core.RequireNotEmpty(t, fixturePath)
func RequireNotEmpty(t testing.TB, v any, msg ...string) {
	t.Helper()
	if assertIsEmpty(v) {
		t.Fatalf("RequireNotEmpty want=non-empty got=%#v%s", v, assertMsg(msg))
	}
}

// AssertElementsMatch fails the test if want and got are not slices/arrays
// containing the same elements regardless of order. Uses deep equality
// per element.
//
//	core.AssertElementsMatch(t, []int{1, 2, 3}, []int{3, 1, 2})
func AssertElementsMatch(t testing.TB, want, got any, msg ...string) {
	t.Helper()
	wv := reflect.ValueOf(want)
	gv := reflect.ValueOf(got)
	if (wv.Kind() != reflect.Slice && wv.Kind() != reflect.Array) ||
		(gv.Kind() != reflect.Slice && gv.Kind() != reflect.Array) {
		t.Errorf("AssertElementsMatch both args must be slices/arrays, got want=%s got=%s%s", wv.Kind(), gv.Kind(), assertMsg(msg))
		return
	}
	if wv.Len() != gv.Len() {
		t.Errorf("AssertElementsMatch len-mismatch want=%d got=%d%s", wv.Len(), gv.Len(), assertMsg(msg))
		return
	}
	matched := make([]bool, gv.Len())
	for i := 0; i < wv.Len(); i++ {
		w := wv.Index(i).Interface()
		found := false
		for j := 0; j < gv.Len(); j++ {
			if matched[j] {
				continue
			}
			if reflect.DeepEqual(w, gv.Index(j).Interface()) {
				matched[j] = true
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AssertElementsMatch missing element=%#v from got=%#v%s", w, got, assertMsg(msg))
			return
		}
	}
}

// assertMsg returns " — <joined msg>" or "" depending on optional context.
func assertMsg(msg []string) string {
	if len(msg) == 0 {
		return ""
	}
	return " — " + Join(" ", msg...)
}

// assertContains is the internal helper for AssertContains.
func assertContains(haystack, needle any) bool {
	hv := reflect.ValueOf(haystack)
	switch hv.Kind() {
	case reflect.String:
		nv := reflect.ValueOf(needle)
		if nv.Kind() != reflect.String {
			return false
		}
		return Contains(hv.String(), nv.String())
	case reflect.Slice, reflect.Array:
		for i := 0; i < hv.Len(); i++ {
			if reflect.DeepEqual(hv.Index(i).Interface(), needle) {
				return true
			}
		}
	case reflect.Map:
		for _, k := range hv.MapKeys() {
			if reflect.DeepEqual(k.Interface(), needle) {
				return true
			}
		}
	}
	return false
}

// assertIsNil reports whether v is a nil value, including typed-nil
// interface values like a (*Foo)(nil) inside an any.
func assertIsNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	}
	return false
}

// assertIsEmpty reports whether v is empty for AssertEmpty/AssertNotEmpty.
// Strings/slices/arrays/maps/channels with len==0 are empty; nil is empty;
// other types use their zero value via reflect.
func assertIsEmpty(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return true
		}
		return assertIsEmpty(rv.Elem().Interface())
	}
	zero := reflect.Zero(rv.Type()).Interface()
	return reflect.DeepEqual(v, zero)
}

// assertCompare returns -1, 0, +1 if a is less, equal, or greater than b.
// The boolean is false when the values are not comparable as a numeric or
// string pair.
func assertCompare(a, b any) (int, bool) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return assertCmpInt64(av.Int(), bv.Int()), true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			ai := av.Int()
			bi := bv.Uint()
			if ai < 0 {
				return -1, true
			}
			return assertCmpUint64(uint64(ai), bi), true
		case reflect.Float32, reflect.Float64:
			return assertCmpFloat64(float64(av.Int()), bv.Float()), true
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bi := bv.Int()
			if bi < 0 {
				return 1, true
			}
			return assertCmpUint64(av.Uint(), uint64(bi)), true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return assertCmpUint64(av.Uint(), bv.Uint()), true
		case reflect.Float32, reflect.Float64:
			return assertCmpFloat64(float64(av.Uint()), bv.Float()), true
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return assertCmpFloat64(av.Float(), float64(bv.Int())), true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return assertCmpFloat64(av.Float(), float64(bv.Uint())), true
		case reflect.Float32, reflect.Float64:
			return assertCmpFloat64(av.Float(), bv.Float()), true
		}
	case reflect.String:
		if bv.Kind() == reflect.String {
			a, b := av.String(), bv.String()
			if a < b {
				return -1, true
			}
			if a > b {
				return 1, true
			}
			return 0, true
		}
	}
	return 0, false
}

func assertCmpInt64(a, b int64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func assertCmpUint64(a, b uint64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func assertCmpFloat64(a, b float64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
