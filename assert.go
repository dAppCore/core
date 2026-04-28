// SPDX-License-Identifier: EUPL-1.2

// AX-shaped assertions for the Core test framework.
//
// The Assert* family records a non-fatal failure; the Require* family
// records and stops the test. Both wrap testing.TB so helpers compose
// across Test, Benchmark, and Fuzz contexts.
//
// # Output format
//
// Default is the AX one-line key=value shape — minimalist, grep-able,
// AI-readable:
//
//	fs_test.go:144: AssertEqual want="hello" got="world"
//	fs_test.go:147: AssertNoError got=open foo: no such file or directory
//
// Set core.AssertVerbose = true (e.g. in TestMain or a setup hook) to
// switch to the multi-line "standard output" format more familiar from
// testify and human-driven test triage:
//
//	fs_test.go:144: AssertEqual failed
//	    want: "hello"
//	    got:  "world"
//
// The flag is global to the test process. Pass = silent in either mode
// (Go test default).
//
// Usage
//
//	core.AssertEqual(t, "expected", actual)
//	core.AssertTrue(t, result.OK)
//	core.AssertNoError(t, err)
//	core.AssertContains(t, []string{"a", "b"}, "a")
//	core.RequireNoError(t, c.Fs().Write(p, "data").Error())
//
// Optional trailing string args are joined into a context suffix on
// the failure line.
package core

// AssertVerbose toggles the failure-message format. When false (default)
// each Assert*/Require* emits the AX one-line shape; when true each
// emits a multi-line testify-style block. Flip it in TestMain when a
// human is reading failures and you want the readable format.
//
//	func TestMain(m *testing.M) {
//	    core.AssertVerbose = true
//	    m.Run()
//	}
var AssertVerbose = false

// assertFail centralises every Assert*/Require* failure path. The
// caller passes a kv-pair list of context (e.g. "want", expected,
// "got", actual). When AssertVerbose is false the helper renders one
// AX line; when true a multi-line block.
func assertFail(t TB, fatal bool, name string, msg []string, kvs ...any) {
	t.Helper()

	suffix := assertMsg(msg)

	if AssertVerbose {
		body := name + " failed\n"
		for i := 0; i+1 < len(kvs); i += 2 {
			body += Sprintf("    %s: %#v\n", kvs[i], kvs[i+1])
		}
		if suffix != "" {
			body += "    msg:" + suffix
		}
		if fatal {
			t.Fatal(body)
		} else {
			t.Error(body)
		}
		return
	}

	parts := make([]string, 0, len(kvs)/2)
	for i := 0; i+1 < len(kvs); i += 2 {
		parts = append(parts, Sprintf("%s=%#v", kvs[i], kvs[i+1]))
	}
	line := name
	if len(parts) > 0 {
		line += " " + Join(" ", parts...)
	}
	line += suffix

	if fatal {
		t.Fatal(line)
	} else {
		t.Error(line)
	}
}

// AssertEqual fails the test if want and got are not deeply equal.
//
//	core.AssertEqual(t, "expected", result.Value)
//	core.AssertEqual(t, 42, result.Count)
func AssertEqual(t TB, want, got any, msg ...string) {
	t.Helper()
	if !DeepEqual(want, got) {
		assertFail(t, false, "AssertEqual", msg, "want", want, "got", got)
	}
}

// AssertNotEqual fails the test if want and got are deeply equal.
//
//	core.AssertNotEqual(t, oldValue, newValue)
func AssertNotEqual(t TB, want, got any, msg ...string) {
	t.Helper()
	if DeepEqual(want, got) {
		assertFail(t, false, "AssertNotEqual", msg, "want!=", want, "got", got)
	}
}

// AssertTrue fails the test if condition is false.
//
//	core.AssertTrue(t, result.OK)
//	core.AssertTrue(t, len(items) > 0, "items must not be empty")
func AssertTrue(t TB, condition bool, msg ...string) {
	t.Helper()
	if !condition {
		assertFail(t, false, "AssertTrue", msg, "want", true, "got", false)
	}
}

// AssertFalse fails the test if condition is true.
//
//	core.AssertFalse(t, c.Fs().Exists(deletedPath))
func AssertFalse(t TB, condition bool, msg ...string) {
	t.Helper()
	if condition {
		assertFail(t, false, "AssertFalse", msg, "want", false, "got", true)
	}
}

// AssertNil fails the test if v is non-nil. Handles typed-nil interface
// values (e.g. a (*Foo)(nil) inside an any).
//
//	core.AssertNil(t, returnedPointer)
func AssertNil(t TB, v any, msg ...string) {
	t.Helper()
	if !assertIsNil(v) {
		assertFail(t, false, "AssertNil", msg, "want", nil, "got", v)
	}
}

// AssertNotNil fails the test if v is nil.
//
//	core.AssertNotNil(t, c.Fs())
func AssertNotNil(t TB, v any, msg ...string) {
	t.Helper()
	if assertIsNil(v) {
		assertFail(t, false, "AssertNotNil", msg, "want", "non-nil", "got", nil)
	}
}

// AssertNoError fails the test if err is non-nil.
//
//	core.AssertNoError(t, c.Fs().Write(p, "data").Error())
func AssertNoError(t TB, err error, msg ...string) {
	t.Helper()
	if err != nil {
		assertFail(t, false, "AssertNoError", msg, "got", err)
	}
}

// AssertError fails the test if err is nil. Optional substring matches
// against err.Error() for tighter assertions.
//
//	core.AssertError(t, parseFails())
//	core.AssertError(t, parseFails(), "invalid syntax")
func AssertError(t TB, err error, msg ...string) {
	t.Helper()
	if err == nil {
		assertFail(t, false, "AssertError", msg, "want", "non-nil", "got", nil)
		return
	}
	for _, want := range msg {
		if !Contains(err.Error(), want) {
			assertFail(t, false, "AssertError", nil, "want-substring", want, "got", err.Error())
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
func AssertContains(t TB, haystack, needle any, msg ...string) {
	t.Helper()
	if !assertContains(haystack, needle) {
		assertFail(t, false, "AssertContains", msg, "haystack", haystack, "needle", needle)
	}
}

// AssertNotContains fails the test if needle IS present in haystack.
func AssertNotContains(t TB, haystack, needle any, msg ...string) {
	t.Helper()
	if assertContains(haystack, needle) {
		assertFail(t, false, "AssertNotContains", msg, "haystack", haystack, "needle (found)", needle)
	}
}

// AssertLen fails the test if the length of v does not equal want.
// Works on strings, slices, arrays, maps, and channels.
//
//	core.AssertLen(t, items, 3)
func AssertLen(t TB, v any, want int, msg ...string) {
	t.Helper()
	rv := ValueOf(v)
	switch rv.Kind() {
	case KindString, KindSlice, KindArray, KindMap, KindChan:
		if rv.Len() != want {
			assertFail(t, false, "AssertLen", msg, "want", want, "got", rv.Len())
		}
	default:
		assertFail(t, false, "AssertLen", msg, "unsupported kind", rv.Kind().String())
	}
}

// AssertEmpty fails the test if v is not empty. Treats zero-length
// strings/slices/arrays/maps/channels and nil as empty.
//
//	core.AssertEmpty(t, results)
func AssertEmpty(t TB, v any, msg ...string) {
	t.Helper()
	if !assertIsEmpty(v) {
		assertFail(t, false, "AssertEmpty", msg, "want", "empty", "got", v)
	}
}

// AssertNotEmpty fails the test if v is empty (see AssertEmpty).
//
//	core.AssertNotEmpty(t, response.Body)
func AssertNotEmpty(t TB, v any, msg ...string) {
	t.Helper()
	if assertIsEmpty(v) {
		assertFail(t, false, "AssertNotEmpty", msg, "want", "non-empty", "got", v)
	}
}

// AssertGreater fails the test if got is not strictly greater than want.
// Works on numeric kinds (int, uint, float) and strings.
//
//	core.AssertGreater(t, count, 0)
func AssertGreater(t TB, got, want any, msg ...string) {
	t.Helper()
	cmp, ok := assertCompare(got, want)
	if !ok {
		assertFail(t, false, "AssertGreater", msg, "incomparable got", got, "want", want)
		return
	}
	if cmp <= 0 {
		assertFail(t, false, "AssertGreater", msg, "got", got, "want>", want)
	}
}

// AssertGreaterOrEqual fails the test if got is less than want.
//
//	core.AssertGreaterOrEqual(t, elapsed, minDuration)
func AssertGreaterOrEqual(t TB, got, want any, msg ...string) {
	t.Helper()
	cmp, ok := assertCompare(got, want)
	if !ok {
		assertFail(t, false, "AssertGreaterOrEqual", msg, "incomparable got", got, "want", want)
		return
	}
	if cmp < 0 {
		assertFail(t, false, "AssertGreaterOrEqual", msg, "got", got, "want>=", want)
	}
}

// AssertLess fails the test if got is not strictly less than want.
//
//	core.AssertLess(t, errorCount, limit)
func AssertLess(t TB, got, want any, msg ...string) {
	t.Helper()
	cmp, ok := assertCompare(got, want)
	if !ok {
		assertFail(t, false, "AssertLess", msg, "incomparable got", got, "want", want)
		return
	}
	if cmp >= 0 {
		assertFail(t, false, "AssertLess", msg, "got", got, "want<", want)
	}
}

// AssertLessOrEqual fails the test if got is greater than want.
func AssertLessOrEqual(t TB, got, want any, msg ...string) {
	t.Helper()
	cmp, ok := assertCompare(got, want)
	if !ok {
		assertFail(t, false, "AssertLessOrEqual", msg, "incomparable got", got, "want", want)
		return
	}
	if cmp > 0 {
		assertFail(t, false, "AssertLessOrEqual", msg, "got", got, "want<=", want)
	}
}

// AssertPanics fails the test if calling fn does not panic.
//
//	core.AssertPanics(t, func() { mustParse("garbage") })
func AssertPanics(t TB, fn func(), msg ...string) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			assertFail(t, false, "AssertPanics", msg, "want", "panic", "got", "normal-return")
		}
	}()
	fn()
}

// AssertNotPanics fails the test if calling fn panics.
//
//	core.AssertNotPanics(t, func() { safeDivide(10, 2) })
func AssertNotPanics(t TB, fn func(), msg ...string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			assertFail(t, false, "AssertNotPanics", msg, "want", "normal-return", "got panic", r)
		}
	}()
	fn()
}

// AssertPanicsWithError fails the test if fn does not panic, or panics
// with a value whose error string does not contain wantSubstr. Argument
// order matches testify's PanicsWithError(t, errString, fn).
//
//	core.AssertPanicsWithError(t, "empty input", func() { mustParse("") })
func AssertPanicsWithError(t TB, wantSubstr string, fn func(), msg ...string) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			assertFail(t, false, "AssertPanicsWithError", msg, "want", "panic", "got", "normal-return")
			return
		}
		var got string
		if err, ok := r.(error); ok {
			got = err.Error()
		} else {
			got = Sprintf("%v", r)
		}
		if !Contains(got, wantSubstr) {
			assertFail(t, false, "AssertPanicsWithError", msg, "want-substring", wantSubstr, "got", got)
		}
	}()
	fn()
}

// AssertErrorIs fails the test if err does not wrap target (Is).
//
//	core.AssertErrorIs(t, err, fs.ErrNotExist)
func AssertErrorIs(t TB, err, target error, msg ...string) {
	t.Helper()
	if !Is(err, target) {
		assertFail(t, false, "AssertErrorIs", msg, "want-target", target, "got", err)
	}
}

// AssertInDelta fails the test if |got - want| > delta. Use for float
// comparisons where exact equality is not appropriate.
//
//	core.AssertInDelta(t, expected, actual, 0.0001)
func AssertInDelta(t TB, want, got, delta float64, msg ...string) {
	t.Helper()
	if IsNaN(want) || IsNaN(got) {
		assertFail(t, false, "AssertInDelta", msg, "NaN involved want", want, "got", got)
		return
	}
	diff := Abs(want - got)
	if diff > delta {
		assertFail(t, false, "AssertInDelta", msg, "want", want, "got", got, "delta", delta, "actual-diff", diff)
	}
}

// AssertSame fails the test if want and got are not the same pointer.
//
//	core.AssertSame(t, c.Fs(), c.Fs())  // singleton check
func AssertSame(t TB, want, got any, msg ...string) {
	t.Helper()
	wv := ValueOf(want)
	gv := ValueOf(got)
	if wv.Kind() != KindPointer || gv.Kind() != KindPointer {
		assertFail(t, false, "AssertSame", msg, "both args must be pointers, want", wv.Kind().String(), "got", gv.Kind().String())
		return
	}
	if wv.Pointer() != gv.Pointer() {
		assertFail(t, false, "AssertSame", msg, "want", Sprintf("%p", want), "got", Sprintf("%p", got))
	}
}

// AssertElementsMatch fails the test if want and got are not slices/arrays
// containing the same elements regardless of order. Uses deep equality
// per element.
//
//	core.AssertElementsMatch(t, []int{1, 2, 3}, []int{3, 1, 2})
func AssertElementsMatch(t TB, want, got any, msg ...string) {
	t.Helper()
	wv := ValueOf(want)
	gv := ValueOf(got)
	if (wv.Kind() != KindSlice && wv.Kind() != KindArray) ||
		(gv.Kind() != KindSlice && gv.Kind() != KindArray) {
		assertFail(t, false, "AssertElementsMatch", msg, "both args must be slices/arrays, want", wv.Kind().String(), "got", gv.Kind().String())
		return
	}
	if wv.Len() != gv.Len() {
		assertFail(t, false, "AssertElementsMatch", msg, "len-mismatch want", wv.Len(), "got", gv.Len())
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
			if DeepEqual(w, gv.Index(j).Interface()) {
				matched[j] = true
				found = true
				break
			}
		}
		if !found {
			assertFail(t, false, "AssertElementsMatch", msg, "missing element", w, "from got", got)
			return
		}
	}
}

// RequireNoError fails the test AND stops it if err is non-nil. Use when
// the rest of the test depends on the operation succeeding.
//
//	core.RequireNoError(t, c.Fs().Write(p, "data").Error())
func RequireNoError(t TB, err error, msg ...string) {
	t.Helper()
	if err != nil {
		assertFail(t, true, "RequireNoError", msg, "got", err)
	}
}

// RequireTrue fails the test AND stops it if condition is false. Use to
// guard test invariants where continuing past a false condition is
// meaningless.
//
//	core.RequireTrue(t, scanner.Scan(), "scan precondition")
func RequireTrue(t TB, condition bool, msg ...string) {
	t.Helper()
	if !condition {
		assertFail(t, true, "RequireTrue", msg, "want", true, "got", false)
	}
}

// RequireNotEmpty fails the test AND stops it if v is empty. Use to
// guard test invariants where downstream assertions presume non-empty
// fixture data.
//
//	core.RequireNotEmpty(t, fixturePath)
func RequireNotEmpty(t TB, v any, msg ...string) {
	t.Helper()
	if assertIsEmpty(v) {
		assertFail(t, true, "RequireNotEmpty", msg, "want", "non-empty", "got", v)
	}
}

// --- internal helpers ---

// assertMsg returns " — <joined msg>" or "" depending on optional context.
func assertMsg(msg []string) string {
	if len(msg) == 0 {
		return ""
	}
	return " — " + Join(" ", msg...)
}

// assertContains is the internal helper for AssertContains.
func assertContains(haystack, needle any) bool {
	hv := ValueOf(haystack)
	switch hv.Kind() {
	case KindString:
		nv := ValueOf(needle)
		if nv.Kind() != KindString {
			return false
		}
		return Contains(hv.String(), nv.String())
	case KindSlice, KindArray:
		for i := 0; i < hv.Len(); i++ {
			if DeepEqual(hv.Index(i).Interface(), needle) {
				return true
			}
		}
	case KindMap:
		for _, k := range hv.MapKeys() {
			if DeepEqual(k.Interface(), needle) {
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
	rv := ValueOf(v)
	switch rv.Kind() {
	case KindChan, KindFunc, KindInterface, KindMap, KindPointer, KindSlice:
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
	rv := ValueOf(v)
	switch rv.Kind() {
	case KindString, KindSlice, KindArray, KindMap, KindChan:
		return rv.Len() == 0
	case KindPointer, KindInterface:
		if rv.IsNil() {
			return true
		}
		return assertIsEmpty(rv.Elem().Interface())
	}
	zero := Zero(rv.Type()).Interface()
	return DeepEqual(v, zero)
}

// assertCompare returns -1, 0, +1 if a is less, equal, or greater than b.
// The boolean is false when the values are not comparable as a numeric or
// string pair.
func assertCompare(a, b any) (int, bool) {
	av := ValueOf(a)
	bv := ValueOf(b)
	switch av.Kind() {
	case KindInt, KindInt8, KindInt16, KindInt32, KindInt64:
		switch bv.Kind() {
		case KindInt, KindInt8, KindInt16, KindInt32, KindInt64:
			return assertCmpInt64(av.Int(), bv.Int()), true
		case KindUint, KindUint8, KindUint16, KindUint32, KindUint64:
			ai := av.Int()
			bi := bv.Uint()
			if ai < 0 {
				return -1, true
			}
			return assertCmpUint64(uint64(ai), bi), true
		case KindFloat32, KindFloat64:
			return assertCmpFloat64(float64(av.Int()), bv.Float()), true
		}
	case KindUint, KindUint8, KindUint16, KindUint32, KindUint64:
		switch bv.Kind() {
		case KindInt, KindInt8, KindInt16, KindInt32, KindInt64:
			bi := bv.Int()
			if bi < 0 {
				return 1, true
			}
			return assertCmpUint64(av.Uint(), uint64(bi)), true
		case KindUint, KindUint8, KindUint16, KindUint32, KindUint64:
			return assertCmpUint64(av.Uint(), bv.Uint()), true
		case KindFloat32, KindFloat64:
			return assertCmpFloat64(float64(av.Uint()), bv.Float()), true
		}
	case KindFloat32, KindFloat64:
		switch bv.Kind() {
		case KindInt, KindInt8, KindInt16, KindInt32, KindInt64:
			return assertCmpFloat64(av.Float(), float64(bv.Int())), true
		case KindUint, KindUint8, KindUint16, KindUint32, KindUint64:
			return assertCmpFloat64(av.Float(), float64(bv.Uint())), true
		case KindFloat32, KindFloat64:
			return assertCmpFloat64(av.Float(), bv.Float()), true
		}
	case KindString:
		if bv.Kind() == KindString {
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
