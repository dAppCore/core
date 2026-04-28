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

func (s *stubT) Helper() {}
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
