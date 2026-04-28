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

func TestArray_NewArray_Good(t *T) {
	agents := NewArray("codex", "hades", "homelab")

	AssertEqual(t, 3, agents.Len())
	AssertTrue(t, agents.Contains("hades"))
}

func TestArray_NewArray_Bad(t *T) {
	agents := NewArray[string]()

	AssertEqual(t, 0, agents.Len())
	AssertNil(t, agents.AsSlice())
}

func TestArray_NewArray_Ugly(t *T) {
	agents := NewArray("codex", "codex")

	AssertEqual(t, 2, agents.Len())
	agents.Deduplicate()
	AssertEqual(t, []string{"codex"}, agents.AsSlice())
}

func TestArray_Array_Add_Good(t *T) {
	agents := NewArray("codex")
	agents.Add("hades", "homelab")

	AssertEqual(t, []string{"codex", "hades", "homelab"}, agents.AsSlice())
}

func TestArray_Array_Add_Bad(t *T) {
	agents := NewArray("codex")
	agents.Add()

	AssertEqual(t, []string{"codex"}, agents.AsSlice())
}

func TestArray_Array_Add_Ugly(t *T) {
	agents := NewArray[string]()
	agents.Add("")

	AssertEqual(t, []string{""}, agents.AsSlice())
}

func TestArray_Array_AddUnique_Good(t *T) {
	agents := NewArray("codex", "hades")
	agents.AddUnique("hades", "homelab")

	AssertEqual(t, []string{"codex", "hades", "homelab"}, agents.AsSlice())
}

func TestArray_Array_AddUnique_Bad(t *T) {
	agents := NewArray("codex")
	agents.AddUnique("codex", "codex")

	AssertEqual(t, []string{"codex"}, agents.AsSlice())
}

func TestArray_Array_AddUnique_Ugly(t *T) {
	agents := NewArray("")
	agents.AddUnique("", "codex")

	AssertEqual(t, []string{"", "codex"}, agents.AsSlice())
}

func TestArray_Array_AsSlice_Good(t *T) {
	agents := NewArray("codex", "hades")

	AssertEqual(t, []string{"codex", "hades"}, agents.AsSlice())
}

func TestArray_Array_AsSlice_Bad(t *T) {
	agents := NewArray[string]()

	AssertNil(t, agents.AsSlice())
}

func TestArray_Array_AsSlice_Ugly(t *T) {
	agents := NewArray("codex")
	copy := agents.AsSlice()
	copy[0] = "hades"

	AssertEqual(t, []string{"codex"}, agents.AsSlice())
	AssertEqual(t, []string{"hades"}, copy)
}

func TestArray_Array_Clear_Good(t *T) {
	agents := NewArray("codex", "hades")
	agents.Clear()

	AssertEqual(t, 0, agents.Len())
	AssertNil(t, agents.AsSlice())
}

func TestArray_Array_Clear_Bad(t *T) {
	agents := NewArray[string]()
	agents.Clear()

	AssertEqual(t, 0, agents.Len())
}

func TestArray_Array_Clear_Ugly(t *T) {
	agents := NewArray("codex")
	agents.Clear()
	agents.Add("hades")

	AssertEqual(t, []string{"hades"}, agents.AsSlice())
}

func TestArray_Array_Contains_Good(t *T) {
	agents := NewArray("codex", "hades")

	AssertTrue(t, agents.Contains("codex"))
}

func TestArray_Array_Contains_Bad(t *T) {
	agents := NewArray("codex", "hades")

	AssertFalse(t, agents.Contains("homelab"))
}

func TestArray_Array_Contains_Ugly(t *T) {
	var agents Array[string]

	AssertFalse(t, agents.Contains(""))
}

func TestArray_Array_Deduplicate_Good(t *T) {
	agents := NewArray("codex", "hades", "codex")
	agents.Deduplicate()

	AssertEqual(t, []string{"codex", "hades"}, agents.AsSlice())
}

func TestArray_Array_Deduplicate_Bad(t *T) {
	agents := NewArray("codex", "hades")
	agents.Deduplicate()

	AssertEqual(t, []string{"codex", "hades"}, agents.AsSlice())
}

func TestArray_Array_Deduplicate_Ugly(t *T) {
	agents := NewArray[string]()
	agents.Deduplicate()

	AssertEmpty(t, agents.AsSlice())
}

func TestArray_Array_Each_Good(t *T) {
	agents := NewArray("codex", "hades")
	seen := []string{}
	agents.Each(func(name string) { seen = append(seen, name) })

	AssertEqual(t, []string{"codex", "hades"}, seen)
}

func TestArray_Array_Each_Bad(t *T) {
	agents := NewArray("codex")

	AssertPanics(t, func() { agents.Each(nil) })
}

func TestArray_Array_Each_Ugly(t *T) {
	agents := NewArray[string]()

	AssertNotPanics(t, func() { agents.Each(nil) })
}

func TestArray_Array_Filter_Good(t *T) {
	agents := NewArray("codex", "hades", "homelab")
	r := agents.Filter(func(name string) bool { return HasPrefix(name, "h") })

	AssertTrue(t, r.OK)
	AssertEqual(t, []string{"hades", "homelab"}, r.Value.(*Array[string]).AsSlice())
}

func TestArray_Array_Filter_Bad(t *T) {
	agents := NewArray("codex")

	AssertPanics(t, func() { agents.Filter(nil) })
}

func TestArray_Array_Filter_Ugly(t *T) {
	agents := NewArray[string]()
	r := agents.Filter(nil)

	AssertTrue(t, r.OK)
	AssertNil(t, r.Value.(*Array[string]).AsSlice())
}

func TestArray_Array_Len_Good(t *T) {
	agents := NewArray("codex", "hades")

	AssertEqual(t, 2, agents.Len())
}

func TestArray_Array_Len_Bad(t *T) {
	agents := NewArray[string]()

	AssertEqual(t, 0, agents.Len())
}

func TestArray_Array_Len_Ugly(t *T) {
	agents := NewArray("codex")
	agents.Add("hades")
	agents.Remove("codex")

	AssertEqual(t, 1, agents.Len())
}

func TestArray_Array_Remove_Good(t *T) {
	agents := NewArray("codex", "hades")
	agents.Remove("codex")

	AssertEqual(t, []string{"hades"}, agents.AsSlice())
}

func TestArray_Array_Remove_Bad(t *T) {
	agents := NewArray("codex")
	agents.Remove("missing")

	AssertEqual(t, []string{"codex"}, agents.AsSlice())
}

func TestArray_Array_Remove_Ugly(t *T) {
	agents := NewArray("codex", "codex", "hades")
	agents.Remove("codex")

	AssertEqual(t, []string{"codex", "hades"}, agents.AsSlice())
}
