// SPDX-License-Identifier: EUPL-1.2

package core

func TestAssert_assertCmpFloat64_Good(t *T) {
	AssertEqual(t, -1, assertCmpFloat64(1.25, 2.5))
}
func TestAssert_assertCmpFloat64_Bad(t *T) {
	AssertEqual(t, 0, assertCmpFloat64(2.5, 2.5))
}
func TestAssert_assertCmpFloat64_Ugly(t *T) {
	AssertEqual(t, 1, assertCmpFloat64(-0.5, -1.5))
}
func TestAssert_assertCmpInt64_Good(t *T) {
	AssertEqual(t, -1, assertCmpInt64(-1, 1))
}
func TestAssert_assertCmpInt64_Bad(t *T) {
	AssertEqual(t, 0, assertCmpInt64(42, 42))
}
func TestAssert_assertCmpInt64_Ugly(t *T) {
	AssertEqual(t, 1, assertCmpInt64(1<<62, -1<<62))
}
func TestAssert_assertCmpUint64_Good(t *T) {
	AssertEqual(t, -1, assertCmpUint64(1, 2))
}
func TestAssert_assertCmpUint64_Bad(t *T) {
	AssertEqual(t, 0, assertCmpUint64(42, 42))
}
func TestAssert_assertCmpUint64_Ugly(t *T) {
	AssertEqual(t, 1, assertCmpUint64(1<<63, 1))
}
func TestAssert_assertCompare_Good(t *T) {
	cmp, ok := assertCompare("agent", "brain")

	AssertTrue(t, ok)
	AssertEqual(t, -1, cmp)
}
func TestAssert_assertCompare_Bad(t *T) {
	cmp, ok := assertCompare(struct{ Name string }{"agent"}, struct{ Name string }{"agent"})

	AssertFalse(t, ok)
	AssertEqual(t, 0, cmp)
}
func TestAssert_assertCompare_Ugly(t *T) {
	cmp, ok := assertCompare(-1, uint(1))

	AssertTrue(t, ok)
	AssertEqual(t, -1, cmp)
}
func TestAssert_assertContains_Good(t *T) {
	AssertTrue(t, assertContains([]string{"agent", "dispatch"}, "dispatch"))
}
func TestAssert_assertContains_Bad(t *T) {
	AssertFalse(t, assertContains("agent dispatch", "missing"))
}
func TestAssert_assertContains_Ugly(t *T) {
	AssertTrue(t, assertContains(map[string]int{"session": 1}, "session"))
}
func TestAssert_assertIsEmpty_Good(t *T) {
	AssertTrue(t, assertIsEmpty(""))
}
func TestAssert_assertIsEmpty_Bad(t *T) {
	AssertFalse(t, assertIsEmpty("agent"))
}
func TestAssert_assertIsEmpty_Ugly(t *T) {
	names := []string{}

	AssertTrue(t, assertIsEmpty(&names))
}
func TestAssert_assertIsNil_Good(t *T) {
	AssertTrue(t, assertIsNil(nil))
}
func TestAssert_assertIsNil_Bad(t *T) {
	AssertFalse(t, assertIsNil(0))
}
func TestAssert_assertIsNil_Ugly(t *T) {
	var sessions map[string]string

	AssertTrue(t, assertIsNil(sessions))
}
func TestAssert_assertMsg_Good(t *T) {
	AssertEqual(t, " — agent retry", assertMsg([]string{"agent", "retry"}))
}
func TestAssert_assertMsg_Bad(t *T) {
	AssertEqual(t, "", assertMsg(nil))
}
func TestAssert_assertMsg_Ugly(t *T) {
	AssertEqual(t, " — lethean degraded", assertMsg([]string{"lethean", "degraded"}))
}

// --- assertFail: the centralised emitter for Assert* / Require* ---

// stubT records Errorf/Fatal calls without aborting the test, so the
// triplet can verify assertFail's two output formats.
type stubT struct {
	*T
	msgs  []string
	fatal bool
}

func (s *stubT) Helper() { /* no-op helper for testing.TB compatibility */ }
func (s *stubT) Error(args ...any) {
	s.msgs = append(s.msgs, Sprint(args...))
}
func (s *stubT) Errorf(format string, args ...any) {
	s.msgs = append(s.msgs, Sprintf(format, args...))
}
func (s *stubT) Fatal(args ...any) {
	s.fatal = true
	s.msgs = append(s.msgs, Sprint(args...))
}
func (s *stubT) Fatalf(format string, args ...any) {
	s.fatal = true
	s.msgs = append(s.msgs, Sprintf(format, args...))
}

func TestAssert_assertFail_Good(t *T) {
	prev := AssertVerbose
	defer func() { AssertVerbose = prev }()
	AssertVerbose = false

	st := &stubT{T: t}
	assertFail(st, false, "TestAssertion", nil, "want", 1, "got", 2)

	AssertEqual(t, false, st.fatal)
	AssertLen(t, st.msgs, 1)
	AssertContains(t, st.msgs[0], "TestAssertion")
	AssertContains(t, st.msgs[0], "want=1")
	AssertContains(t, st.msgs[0], "got=2")
}

func TestAssert_assertFail_Bad(t *T) {
	prev := AssertVerbose
	defer func() { AssertVerbose = prev }()
	AssertVerbose = false

	st := &stubT{T: t}
	assertFail(st, true, "RequireFatal", []string{"agent", "context"}, "want", "non-nil", "got", nil)

	AssertEqual(t, true, st.fatal)
	AssertContains(t, st.msgs[0], "RequireFatal")
	AssertContains(t, st.msgs[0], "agent context")
}

func TestAssert_assertFail_Ugly(t *T) {
	prev := AssertVerbose
	defer func() { AssertVerbose = prev }()
	AssertVerbose = true

	st := &stubT{T: t}
	assertFail(st, false, "VerboseAssertion", []string{"homelab", "drained"}, "want", 42, "got", 13)

	AssertEqual(t, false, st.fatal)
	AssertContains(t, st.msgs[0], "VerboseAssertion failed")
	AssertContains(t, st.msgs[0], "want: 42")
	AssertContains(t, st.msgs[0], "got: 13")
	AssertContains(t, st.msgs[0], "msg:")
}

func TestAssert_assertFailVerboseFatal_Bad(t *T) {
	prev := AssertVerbose
	defer func() { AssertVerbose = prev }()
	AssertVerbose = true

	st := &stubT{T: t}
	assertFail(st, true, "VerboseRequire", []string{"agent", "halted"}, "want", true, "got", false)

	AssertTrue(t, st.fatal)
	AssertContains(t, st.msgs[0], "VerboseRequire failed")
	AssertContains(t, st.msgs[0], "agent halted")
}

func assertStub(t *T) *stubT {
	t.Helper()
	return &stubT{T: t}
}

func assertOneMessage(t *T, st *stubT, want string) {
	t.Helper()
	AssertLen(t, st.msgs, 1)
	AssertContains(t, st.msgs[0], want)
}

func TestAssert_AssertEqualityFailures_Good(t *T) {
	st := assertStub(t)
	AssertEqual(st, "expected", "actual")
	assertOneMessage(t, st, "AssertEqual")

	st = assertStub(t)
	AssertNotEqual(st, "same", "same")
	assertOneMessage(t, st, "AssertNotEqual")
}

func TestAssert_AssertBooleanNilFailures_Bad(t *T) {
	st := assertStub(t)
	AssertTrue(st, false)
	assertOneMessage(t, st, "AssertTrue")

	st = assertStub(t)
	AssertFalse(st, true)
	assertOneMessage(t, st, "AssertFalse")

	st = assertStub(t)
	AssertNil(st, 42)
	assertOneMessage(t, st, "AssertNil")

	st = assertStub(t)
	AssertNotNil(st, nil)
	assertOneMessage(t, st, "AssertNotNil")

	st = assertStub(t)
	AssertNoError(st, AnError)
	assertOneMessage(t, st, "AssertNoError")

	st = assertStub(t)
	AssertError(st, nil)
	assertOneMessage(t, st, "AssertError")
}

func TestAssert_AssertCollectionFailures_Ugly(t *T) {
	st := assertStub(t)
	AssertError(st, AnError, "missing")
	assertOneMessage(t, st, "want-substring")

	st = assertStub(t)
	AssertContains(st, "agent", "missing")
	assertOneMessage(t, st, "AssertContains")

	st = assertStub(t)
	AssertNotContains(st, "agent", "gen")
	assertOneMessage(t, st, "AssertNotContains")

	st = assertStub(t)
	AssertLen(st, []string{"agent"}, 2)
	assertOneMessage(t, st, "AssertLen")

	st = assertStub(t)
	AssertLen(st, 42, 1)
	assertOneMessage(t, st, "unsupported kind")

	st = assertStub(t)
	AssertEmpty(st, "agent")
	assertOneMessage(t, st, "AssertEmpty")

	st = assertStub(t)
	AssertNotEmpty(st, "")
	assertOneMessage(t, st, "AssertNotEmpty")
}

func TestAssert_AssertOrderingFailures_Good(t *T) {
	st := assertStub(t)
	AssertGreater(st, struct{}{}, struct{}{})
	assertOneMessage(t, st, "incomparable got")

	st = assertStub(t)
	AssertGreater(st, 1, 2)
	assertOneMessage(t, st, "AssertGreater")

	st = assertStub(t)
	AssertGreaterOrEqual(st, struct{}{}, struct{}{})
	assertOneMessage(t, st, "AssertGreaterOrEqual")

	st = assertStub(t)
	AssertGreaterOrEqual(st, 1, 2)
	assertOneMessage(t, st, "want>=")

	st = assertStub(t)
	AssertLess(st, struct{}{}, struct{}{})
	assertOneMessage(t, st, "AssertLess")

	st = assertStub(t)
	AssertLess(st, 2, 1)
	assertOneMessage(t, st, "want<")

	st = assertStub(t)
	AssertLessOrEqual(st, struct{}{}, struct{}{})
	assertOneMessage(t, st, "AssertLessOrEqual")

	st = assertStub(t)
	AssertLessOrEqual(st, 2, 1)
	assertOneMessage(t, st, "want<=")
}

func TestAssert_AssertPanicFailures_Bad(t *T) {
	st := assertStub(t)
	AssertPanics(st, func() {})
	assertOneMessage(t, st, "normal-return")

	st = assertStub(t)
	AssertNotPanics(st, func() { panic("agent failed") })
	assertOneMessage(t, st, "got panic")

	st = assertStub(t)
	AssertPanicsWithError(st, "agent failed", func() {})
	assertOneMessage(t, st, "normal-return")

	st = assertStub(t)
	AssertPanicsWithError(st, "agent failed", func() { panic("other failure") })
	assertOneMessage(t, st, "want-substring")
}

func TestAssert_AssertPointerSliceFailures_Ugly(t *T) {
	st := assertStub(t)
	AssertErrorIs(st, AnError, NewError("different"))
	assertOneMessage(t, st, "AssertErrorIs")

	st = assertStub(t)
	AssertInDelta(st, NaN(), 1, 0.1)
	assertOneMessage(t, st, "NaN involved want")

	st = assertStub(t)
	AssertInDelta(st, 1, 2, 0.1)
	assertOneMessage(t, st, "actual-diff")

	left := 1
	right := 1

	st = assertStub(t)
	AssertSame(st, 1, 1)
	assertOneMessage(t, st, "both args must be pointers")

	st = assertStub(t)
	AssertSame(st, &left, &right)
	assertOneMessage(t, st, "AssertSame")

	st = assertStub(t)
	AssertElementsMatch(st, 1, []int{1})
	assertOneMessage(t, st, "both args must be slices")

	st = assertStub(t)
	AssertElementsMatch(st, []int{1}, []int{1, 2})
	assertOneMessage(t, st, "len-mismatch")

	st = assertStub(t)
	AssertElementsMatch(st, []int{1, 3}, []int{1, 2})
	assertOneMessage(t, st, "missing element")
}

func TestAssert_RequireFailures_Good(t *T) {
	st := assertStub(t)
	RequireNoError(st, AnError)
	assertOneMessage(t, st, "RequireNoError")
	AssertTrue(t, st.fatal)

	st = assertStub(t)
	RequireTrue(st, false)
	assertOneMessage(t, st, "RequireTrue")
	AssertTrue(t, st.fatal)

	st = assertStub(t)
	RequireNotEmpty(st, "")
	assertOneMessage(t, st, "RequireNotEmpty")
	AssertTrue(t, st.fatal)
}

func TestAssert_assertContainsStringNeedle_Bad(t *T) {
	AssertFalse(t, assertContains("agent", 42))
}

func TestAssert_assertIsEmptyTypedNil_Ugly(t *T) {
	var names *[]string

	AssertTrue(t, assertIsEmpty(names))
}

func TestAssert_assertCompareMixedKinds_Good(t *T) {
	cmp, ok := assertCompare(int64(4), uint64(3))
	AssertTrue(t, ok)
	AssertEqual(t, 1, cmp)

	cmp, ok = assertCompare(uint(4), int(5))
	AssertTrue(t, ok)
	AssertEqual(t, -1, cmp)

	cmp, ok = assertCompare(uint(4), uint64(4))
	AssertTrue(t, ok)
	AssertEqual(t, 0, cmp)
}

func TestAssert_assertCompareMixedKinds_Bad(t *T) {
	cmp, ok := assertCompare(uint(4), 4.5)
	AssertTrue(t, ok)
	AssertEqual(t, -1, cmp)

	cmp, ok = assertCompare(float64(6), uint(5))
	AssertTrue(t, ok)
	AssertEqual(t, 1, cmp)

	cmp, ok = assertCompare(float32(3.5), float64(3.5))
	AssertTrue(t, ok)
	AssertEqual(t, 0, cmp)
}

func TestAssert_assertCompareMixedKinds_Ugly(t *T) {
	cmp, ok := assertCompare("brain", "agent")
	AssertTrue(t, ok)
	AssertEqual(t, 1, cmp)
}
