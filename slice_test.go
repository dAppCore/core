package core_test

import (
	. "dappco.re/go"
)

func TestSlice_SliceContains_Good(t *T) {
	AssertTrue(t, SliceContains([]string{"a", "b"}, "b"))
}

func TestSlice_SliceContains_Bad(t *T) {
	AssertFalse(t, SliceContains([]string{"a", "b"}, "c"))
}

func TestSlice_SliceContains_Ugly(t *T) {
	AssertFalse(t, SliceContains([]int(nil), 0))
}

func TestSlice_SliceIndex_Good(t *T) {
	AssertEqual(t, 1, SliceIndex([]string{"a", "b"}, "b"))
}

func TestSlice_SliceIndex_Bad(t *T) {
	AssertEqual(t, -1, SliceIndex([]string{"a", "b"}, "c"))
}

func TestSlice_SliceIndex_Ugly(t *T) {
	AssertEqual(t, 0, SliceIndex([]int{7, 7, 7}, 7))
}

func TestSlice_SliceSort_Good(t *T) {
	items := []int{3, 1, 2}
	SliceSort(items)

	AssertEqual(t, []int{1, 2, 3}, items)
}

func TestSlice_SliceSort_Bad(t *T) {
	var items []int
	SliceSort(items)

	AssertNil(t, items)
}

func TestSlice_SliceSort_Ugly(t *T) {
	items := []string{"beta", "alpha", "beta"}
	SliceSort(items)

	AssertEqual(t, []string{"alpha", "beta", "beta"}, items)
}

func TestSlice_SliceUniq_Good(t *T) {
	items := SliceUniq([]string{"a", "b", "a"})

	AssertEqual(t, []string{"a", "b"}, items)
}

func TestSlice_SliceUniq_Bad(t *T) {
	AssertNil(t, SliceUniq([]string(nil)))
}

func TestSlice_SliceUniq_Ugly(t *T) {
	items := []int{3, 3, 2, 1, 2, 1}

	AssertEqual(t, []int{3, 2, 1}, SliceUniq(items))
}

func TestSlice_SliceReverse_Good(t *T) {
	items := []int{1, 2, 3}
	SliceReverse(items)

	AssertEqual(t, []int{3, 2, 1}, items)
}

func TestSlice_SliceReverse_Bad(t *T) {
	var items []int
	SliceReverse(items)

	AssertNil(t, items)
}

func TestSlice_SliceReverse_Ugly(t *T) {
	items := []string{"a", "b", "c", "d"}
	SliceReverse(items)

	AssertEqual(t, []string{"d", "c", "b", "a"}, items)
}

func TestSlice_SliceAll_Good(t *T) {
	AssertTrue(t, SliceAll([]string{"agent", "agent.dispatch"}, func(name string) bool {
		return HasPrefix(name, "agent")
	}))
}

func TestSlice_SliceAll_Bad(t *T) {
	AssertFalse(t, SliceAll([]int{2, 4, 5}, func(n int) bool { return n%2 == 0 }))
}

func TestSlice_SliceAll_Ugly(t *T) {
	AssertTrue(t, SliceAll([]string{}, func(string) bool { return false }))
}

func TestSlice_SliceAny_Good(t *T) {
	AssertTrue(t, SliceAny([]string{"codex", "hades"}, func(name string) bool {
		return name == "hades"
	}))
}

func TestSlice_SliceAny_Bad(t *T) {
	AssertFalse(t, SliceAny([]string{"codex", "hades"}, func(name string) bool {
		return name == "homelab"
	}))
}

func TestSlice_SliceAny_Ugly(t *T) {
	AssertFalse(t, SliceAny([]string{}, nil))
}

func TestSlice_SliceClone_Good(t *T) {
	clone := SliceClone([]string{"codex", "hades"})

	AssertEqual(t, []string{"codex", "hades"}, clone)
}

func TestSlice_SliceClone_Bad(t *T) {
	AssertNil(t, SliceClone[string](nil))
}

func TestSlice_SliceClone_Ugly(t *T) {
	original := []string{"codex"}
	clone := SliceClone(original)
	clone[0] = "hades"

	AssertEqual(t, []string{"codex"}, original)
	AssertEqual(t, []string{"hades"}, clone)
}

func TestSlice_SliceDrop_Good(t *T) {
	AssertEqual(t, []string{"dispatch", "ready"}, SliceDrop([]string{"agent", "dispatch", "ready"}, 1))
}

func TestSlice_SliceDrop_Bad(t *T) {
	AssertNil(t, SliceDrop([]string{"agent"}, 2))
}

func TestSlice_SliceDrop_Ugly(t *T) {
	items := []string{"agent", "dispatch"}

	AssertEqual(t, items, SliceDrop(items, 0))
}

func TestSlice_SliceFilter_Good(t *T) {
	got := SliceFilter([]string{"codex", "hades", "homelab"}, func(name string) bool {
		return HasPrefix(name, "h")
	})

	AssertEqual(t, []string{"hades", "homelab"}, got)
}

func TestSlice_SliceFilter_Bad(t *T) {
	got := SliceFilter([]string{"codex", "hades"}, func(name string) bool {
		return name == "missing"
	})

	AssertEmpty(t, got)
}

func TestSlice_SliceFilter_Ugly(t *T) {
	AssertNil(t, SliceFilter([]string(nil), func(string) bool { return true }))
}

func TestSlice_SliceFlatMap_Good(t *T) {
	got := SliceFlatMap([]string{"agent dispatch", "homelab ready"}, func(line string) []string {
		return Split(line, " ")
	})

	AssertEqual(t, []string{"agent", "dispatch", "homelab", "ready"}, got)
}

func TestSlice_SliceFlatMap_Bad(t *T) {
	AssertNil(t, SliceFlatMap([]string{}, func(line string) []string { return []string{line} }))
}

func TestSlice_SliceFlatMap_Ugly(t *T) {
	got := SliceFlatMap([]string{"agent", "", "ready"}, func(line string) []string {
		if line == "" {
			return nil
		}
		return []string{line}
	})

	AssertEqual(t, []string{"agent", "ready"}, got)
}

func TestSlice_SliceMap_Good(t *T) {
	got := SliceMap([]string{"codex", "hades"}, func(name string) string { return Upper(name) })

	AssertEqual(t, []string{"CODEX", "HADES"}, got)
}

func TestSlice_SliceMap_Bad(t *T) {
	AssertNil(t, SliceMap([]string{}, func(name string) string { return Upper(name) }))
}

func TestSlice_SliceMap_Ugly(t *T) {
	got := SliceMap([]string{"", "agent"}, func(name string) int { return RuneCount(name) })

	AssertEqual(t, []int{0, 5}, got)
}

func TestSlice_SliceReduce_Good(t *T) {
	got := SliceReduce([]int{1, 2, 3}, 0, func(total, n int) int { return total + n })

	AssertEqual(t, 6, got)
}

func TestSlice_SliceReduce_Bad(t *T) {
	got := SliceReduce([]int{}, 42, func(total, n int) int { return total + n })

	AssertEqual(t, 42, got)
}

func TestSlice_SliceReduce_Ugly(t *T) {
	got := SliceReduce([]string{"agent", "dispatch"}, "", func(total, part string) string {
		if total == "" {
			return part
		}
		return Join(".", total, part)
	})

	AssertEqual(t, "agent.dispatch", got)
}

func TestSlice_SliceSorted_Good(t *T) {
	seq := func(yield func(string) bool) {
		yield("hades")
		yield("codex")
		yield("homelab")
	}

	AssertEqual(t, []string{"codex", "hades", "homelab"}, SliceSorted(seq))
}

func TestSlice_SliceSorted_Bad(t *T) {
	seq := func(yield func(int) bool) {}

	AssertEmpty(t, SliceSorted(seq))
}

func TestSlice_SliceSorted_Ugly(t *T) {
	seq := func(yield func(int) bool) {
		yield(3)
		yield(1)
		yield(3)
	}

	AssertEqual(t, []int{1, 3, 3}, SliceSorted(seq))
}

func TestSlice_SliceTake_Good(t *T) {
	AssertEqual(t, []string{"agent", "dispatch"}, SliceTake([]string{"agent", "dispatch", "ready"}, 2))
}

func TestSlice_SliceTake_Bad(t *T) {
	AssertNil(t, SliceTake([]string{"agent"}, 0))
}

func TestSlice_SliceTake_Ugly(t *T) {
	items := []string{"agent", "dispatch"}

	AssertEqual(t, items, SliceTake(items, 5))
}
