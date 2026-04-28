package core_test

import (
	. "dappco.re/go"
)

// --- Array[T] ---

func TestArray_New_Good(t *T) {
	a := NewArray("a", "b", "c")
	AssertEqual(t, 3, a.Len())
}

func TestArray_Add_Good(t *T) {
	a := NewArray[string]()
	a.Add("x", "y")
	AssertEqual(t, 2, a.Len())
	AssertTrue(t, a.Contains("x"))
	AssertTrue(t, a.Contains("y"))
}

func TestArray_AddUnique_Good(t *T) {
	a := NewArray("a", "b")
	a.AddUnique("b", "c")
	AssertEqual(t, 3, a.Len())
}

func TestArray_Contains_Good(t *T) {
	a := NewArray(1, 2, 3)
	AssertTrue(t, a.Contains(2))
	AssertFalse(t, a.Contains(99))
}

func TestArray_Filter_Good(t *T) {
	a := NewArray(1, 2, 3, 4, 5)
	r := a.Filter(func(n int) bool { return n%2 == 0 })
	AssertTrue(t, r.OK)
	evens := r.Value.(*Array[int])
	AssertEqual(t, 2, evens.Len())
	AssertTrue(t, evens.Contains(2))
	AssertTrue(t, evens.Contains(4))
}

func TestArray_Each_Good(t *T) {
	a := NewArray("a", "b", "c")
	var collected []string
	a.Each(func(s string) { collected = append(collected, s) })
	AssertEqual(t, []string{"a", "b", "c"}, collected)
}

func TestArray_Remove_Good(t *T) {
	a := NewArray("a", "b", "c")
	a.Remove("b")
	AssertEqual(t, 2, a.Len())
	AssertFalse(t, a.Contains("b"))
}

func TestArray_Remove_Bad(t *T) {
	a := NewArray("a", "b")
	a.Remove("missing")
	AssertEqual(t, 2, a.Len())
}

func TestArray_Deduplicate_Good(t *T) {
	a := NewArray("a", "b", "a", "c", "b")
	a.Deduplicate()
	AssertEqual(t, 3, a.Len())
}

func TestArray_Clear_Good(t *T) {
	a := NewArray(1, 2, 3)
	a.Clear()
	AssertEqual(t, 0, a.Len())
}

func TestArray_AsSlice_Good(t *T) {
	a := NewArray("x", "y")
	s := a.AsSlice()
	AssertEqual(t, []string{"x", "y"}, s)
}

func TestArray_Empty_Good(t *T) {
	a := NewArray[int]()
	AssertEqual(t, 0, a.Len())
	AssertFalse(t, a.Contains(0))
	AssertEqual(t, []int(nil), a.AsSlice())
}
