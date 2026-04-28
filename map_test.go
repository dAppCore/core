package core_test

import (
	. "dappco.re/go"
)

func TestMap_MapKeys_Good(t *T) {
	keys := MapKeys(map[string]int{"a": 1, "b": 2})

	AssertElementsMatch(t, []string{"a", "b"}, keys)
}

func TestMap_MapKeys_Bad(t *T) {
	keys := MapKeys[string, int](nil)

	AssertEmpty(t, keys)
}

func TestMap_MapKeys_Ugly(t *T) {
	keys := MapKeys(map[int]string{2: "b", 1: "a", 3: "c"})

	SliceSort(keys)
	AssertEqual(t, []int{1, 2, 3}, keys)
}

func TestMap_MapValues_Good(t *T) {
	values := MapValues(map[string]int{"a": 1, "b": 2})

	AssertElementsMatch(t, []int{1, 2}, values)
}

func TestMap_MapValues_Bad(t *T) {
	values := MapValues[string, int](nil)

	AssertEmpty(t, values)
}

func TestMap_MapValues_Ugly(t *T) {
	values := MapValues(map[int]string{2: "b", 1: "a", 3: "c"})

	SliceSort(values)
	AssertEqual(t, []string{"a", "b", "c"}, values)
}

func TestMap_MapClone_Good(t *T) {
	clone := MapClone(map[string]int{"a": 1})

	AssertEqual(t, map[string]int{"a": 1}, clone)
}

func TestMap_MapClone_Bad(t *T) {
	clone := MapClone[string, int](nil)

	AssertNil(t, clone)
}

func TestMap_MapClone_Ugly(t *T) {
	original := map[string]int{"a": 1}
	clone := MapClone(original)
	clone["a"] = 2

	AssertEqual(t, 1, original["a"])
	AssertEqual(t, 2, clone["a"])
}

func TestMap_MapFilter_Good(t *T) {
	filtered := MapFilter(map[string]bool{"agent": true, "archive": false}, func(_ string, enabled bool) bool {
		return enabled
	})

	AssertEqual(t, map[string]bool{"agent": true}, filtered)
}

func TestMap_MapFilter_Bad(t *T) {
	filtered := MapFilter(map[string]bool{"agent": false}, func(_ string, enabled bool) bool {
		return enabled
	})

	AssertEmpty(t, filtered)
}

func TestMap_MapFilter_Ugly(t *T) {
	AssertNil(t, MapFilter[string, bool](nil, func(string, bool) bool { return true }))
}

func TestMap_MapHasKey_Good(t *T) {
	AssertTrue(t, MapHasKey(map[string]int{"agent": 1}, "agent"))
}

func TestMap_MapHasKey_Bad(t *T) {
	AssertFalse(t, MapHasKey(map[string]int{"agent": 1}, "missing"))
}

func TestMap_MapHasKey_Ugly(t *T) {
	AssertFalse(t, MapHasKey(map[string]int(nil), "agent"))
}

func TestMap_MapMerge_Good(t *T) {
	merged := MapMerge(map[string]string{"agent": "codex"}, map[string]string{"region": "homelab"})

	AssertEqual(t, map[string]string{"agent": "codex", "region": "homelab"}, merged)
}

func TestMap_MapMerge_Bad(t *T) {
	merged := MapMerge(map[string]string{"agent": "codex"}, map[string]string{"agent": "hades"})

	AssertEqual(t, map[string]string{"agent": "hades"}, merged)
}

func TestMap_MapMerge_Ugly(t *T) {
	merged := MapMerge[string, string](nil, nil)

	AssertEmpty(t, merged)
}
