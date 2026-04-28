package core_test

import (
	"testing"

	. "dappco.re/go/core"
)

func TestSlice_SliceContains_Good(t *testing.T) {
	AssertTrue(t, SliceContains([]string{"a", "b"}, "b"))
}

func TestSlice_SliceContains_Bad(t *testing.T) {
	AssertFalse(t, SliceContains([]string{"a", "b"}, "c"))
}

func TestSlice_SliceContains_Ugly(t *testing.T) {
	AssertFalse(t, SliceContains([]int(nil), 0))
}

func TestSlice_SliceIndex_Good(t *testing.T) {
	AssertEqual(t, 1, SliceIndex([]string{"a", "b"}, "b"))
}

func TestSlice_SliceIndex_Bad(t *testing.T) {
	AssertEqual(t, -1, SliceIndex([]string{"a", "b"}, "c"))
}

func TestSlice_SliceIndex_Ugly(t *testing.T) {
	AssertEqual(t, 0, SliceIndex([]int{7, 7, 7}, 7))
}

func TestSlice_SliceSort_Good(t *testing.T) {
	items := []int{3, 1, 2}
	SliceSort(items)

	AssertEqual(t, []int{1, 2, 3}, items)
}

func TestSlice_SliceSort_Bad(t *testing.T) {
	var items []int
	SliceSort(items)

	AssertNil(t, items)
}

func TestSlice_SliceSort_Ugly(t *testing.T) {
	items := []string{"beta", "alpha", "beta"}
	SliceSort(items)

	AssertEqual(t, []string{"alpha", "beta", "beta"}, items)
}

func TestSlice_SliceUniq_Good(t *testing.T) {
	items := SliceUniq([]string{"a", "b", "a"})

	AssertEqual(t, []string{"a", "b"}, items)
}

func TestSlice_SliceUniq_Bad(t *testing.T) {
	AssertNil(t, SliceUniq([]string(nil)))
}

func TestSlice_SliceUniq_Ugly(t *testing.T) {
	items := []int{3, 3, 2, 1, 2, 1}

	AssertEqual(t, []int{3, 2, 1}, SliceUniq(items))
}

func TestSlice_SliceReverse_Good(t *testing.T) {
	items := []int{1, 2, 3}
	SliceReverse(items)

	AssertEqual(t, []int{3, 2, 1}, items)
}

func TestSlice_SliceReverse_Bad(t *testing.T) {
	var items []int
	SliceReverse(items)

	AssertNil(t, items)
}

func TestSlice_SliceReverse_Ugly(t *testing.T) {
	items := []string{"a", "b", "c", "d"}
	SliceReverse(items)

	AssertEqual(t, []string{"d", "c", "b", "a"}, items)
}
