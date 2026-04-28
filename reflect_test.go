package core_test

import (
	. "dappco.re/go"
)

func TestReflect_DeepEqual_Good(t *T) {
	left := map[string]int{"agent": 1, "health": 2}
	right := map[string]int{"health": 2, "agent": 1}
	AssertTrue(t, DeepEqual(left, right))
}

func TestReflect_DeepEqual_Bad(t *T) {
	AssertFalse(t, DeepEqual([]string{"agent"}, []string{"operator"}))
}

func TestReflect_DeepEqual_Ugly(t *T) {
	AssertFalse(t, DeepEqual([]string(nil), []string{}))
}

func TestReflect_TypeOf_Good(t *T) {
	AssertEqual(t, KindString, TypeOf("agent").Kind())
}

func TestReflect_TypeOf_Bad(t *T) {
	AssertNil(t, TypeOf(nil))
}

func TestReflect_TypeOf_Ugly(t *T) {
	var c *Core
	AssertEqual(t, KindPointer, TypeOf(c).Kind())
}

func TestReflect_ValueOf_Good(t *T) {
	AssertEqual(t, int64(42), ValueOf(42).Int())
}

func TestReflect_ValueOf_Bad(t *T) {
	AssertEqual(t, KindInvalid, ValueOf(nil).Kind())
}

func TestReflect_ValueOf_Ugly(t *T) {
	var c *Core
	v := ValueOf(c)
	AssertEqual(t, KindPointer, v.Kind())
	AssertTrue(t, v.IsNil())
}

func TestReflect_Zero_Good(t *T) {
	z := Zero(TypeOf(42))
	AssertEqual(t, 0, z.Interface())
}

func TestReflect_Zero_Bad(t *T) {
	AssertPanics(t, func() {
		_ = Zero(nil)
	})
}

func TestReflect_Zero_Ugly(t *T) {
	z := Zero(TypeOf((*Core)(nil)))
	AssertEqual(t, KindPointer, z.Kind())
	AssertTrue(t, z.IsNil())
}
