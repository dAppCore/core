// SPDX-License-Identifier: EUPL-1.2

package core

func TestTest_assertCmpFloat64_Good(t *T) {
	AssertEqual(t, -1, assertCmpFloat64(1.25, 2.5))
}
func TestTest_assertCmpFloat64_Bad(t *T) {
	AssertEqual(t, 0, assertCmpFloat64(2.5, 2.5))
}
func TestTest_assertCmpFloat64_Ugly(t *T) {
	AssertEqual(t, 1, assertCmpFloat64(-0.5, -1.5))
}
func TestTest_assertCmpInt64_Good(t *T) {
	AssertEqual(t, -1, assertCmpInt64(-1, 1))
}
func TestTest_assertCmpInt64_Bad(t *T) {
	AssertEqual(t, 0, assertCmpInt64(42, 42))
}
func TestTest_assertCmpInt64_Ugly(t *T) {
	AssertEqual(t, 1, assertCmpInt64(1<<62, -1<<62))
}
func TestTest_assertCmpUint64_Good(t *T) {
	AssertEqual(t, -1, assertCmpUint64(1, 2))
}
func TestTest_assertCmpUint64_Bad(t *T) {
	AssertEqual(t, 0, assertCmpUint64(42, 42))
}
func TestTest_assertCmpUint64_Ugly(t *T) {
	AssertEqual(t, 1, assertCmpUint64(1<<63, 1))
}
func TestTest_assertCompare_Good(t *T) {
	cmp, ok := assertCompare("agent", "brain")

	AssertTrue(t, ok)
	AssertEqual(t, -1, cmp)
}
func TestTest_assertCompare_Bad(t *T) {
	cmp, ok := assertCompare(struct{ Name string }{"agent"}, struct{ Name string }{"agent"})

	AssertFalse(t, ok)
	AssertEqual(t, 0, cmp)
}
func TestTest_assertCompare_Ugly(t *T) {
	cmp, ok := assertCompare(-1, uint(1))

	AssertTrue(t, ok)
	AssertEqual(t, -1, cmp)
}
func TestTest_assertContains_Good(t *T) {
	AssertTrue(t, assertContains([]string{"agent", "dispatch"}, "dispatch"))
}
func TestTest_assertContains_Bad(t *T) {
	AssertFalse(t, assertContains("agent dispatch", "missing"))
}
func TestTest_assertContains_Ugly(t *T) {
	AssertTrue(t, assertContains(map[string]int{"session": 1}, "session"))
}
func TestTest_assertIsEmpty_Good(t *T) {
	AssertTrue(t, assertIsEmpty(""))
}
func TestTest_assertIsEmpty_Bad(t *T) {
	AssertFalse(t, assertIsEmpty("agent"))
}
func TestTest_assertIsEmpty_Ugly(t *T) {
	names := []string{}

	AssertTrue(t, assertIsEmpty(&names))
}
func TestTest_assertIsNil_Good(t *T) {
	AssertTrue(t, assertIsNil(nil))
}
func TestTest_assertIsNil_Bad(t *T) {
	AssertFalse(t, assertIsNil(0))
}
func TestTest_assertIsNil_Ugly(t *T) {
	var sessions map[string]string

	AssertTrue(t, assertIsNil(sessions))
}
func TestTest_assertMsg_Good(t *T) {
	AssertEqual(t, " — agent retry", assertMsg([]string{"agent", "retry"}))
}
func TestTest_assertMsg_Bad(t *T) {
	AssertEqual(t, "", assertMsg(nil))
}
func TestTest_assertMsg_Ugly(t *T) {
	AssertEqual(t, " — lethean degraded", assertMsg([]string{"lethean", "degraded"}))
}
