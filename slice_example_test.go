package core_test

import . "dappco.re/go"

// ExampleSliceContains checks list membership through `SliceContains` for agent lists.
// List transforms, predicates, and reducers remain typed core helpers.
func ExampleSliceContains() {
	Println(SliceContains([]string{"alpha", "bravo"}, "bravo"))
	// Output: true
}

// ExampleSliceIndex finds a list index through `SliceIndex` for agent lists. List
// transforms, predicates, and reducers remain typed core helpers.
func ExampleSliceIndex() {
	Println(SliceIndex([]string{"alpha", "bravo"}, "bravo"))
	// Output: 1
}

// ExampleSliceSort sorts a list through `SliceSort` for agent lists. List transforms,
// predicates, and reducers remain typed core helpers.
func ExampleSliceSort() {
	names := []string{"charlie", "alpha", "bravo"}
	SliceSort(names)
	Println(names)
	// Output: [alpha bravo charlie]
}

// ExampleSliceUniq deduplicates a list through `SliceUniq` for agent lists. List
// transforms, predicates, and reducers remain typed core helpers.
func ExampleSliceUniq() {
	Println(SliceUniq([]string{"alpha", "alpha", "bravo"}))
	// Output: [alpha bravo]
}

// ExampleSliceReverse reverses a list through `SliceReverse` for agent lists. List
// transforms, predicates, and reducers remain typed core helpers.
func ExampleSliceReverse() {
	names := []string{"alpha", "bravo", "charlie"}
	SliceReverse(names)
	Println(names)
	// Output: [charlie bravo alpha]
}

// ExampleSliceFilter filters a list through `SliceFilter` for agent lists. List
// transforms, predicates, and reducers remain typed core helpers.
func ExampleSliceFilter() {
	evens := SliceFilter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
	Println(evens)
	// Output: [2 4]
}

// ExampleSliceMap maps list values through `SliceMap` for agent lists. List transforms,
// predicates, and reducers remain typed core helpers.
func ExampleSliceMap() {
	labels := SliceMap([]int{1, 2, 3}, func(n int) string { return Sprintf("node-%d", n) })
	Println(labels)
	// Output: [node-1 node-2 node-3]
}

// ExampleSliceReduce reduces list values through `SliceReduce` for agent lists. List
// transforms, predicates, and reducers remain typed core helpers.
func ExampleSliceReduce() {
	sum := SliceReduce([]int{1, 2, 3}, 0, func(acc, n int) int { return acc + n })
	Println(sum)
	// Output: 6
}

// ExampleSliceFlatMap expands list values through `SliceFlatMap` for agent lists. List
// transforms, predicates, and reducers remain typed core helpers.
func ExampleSliceFlatMap() {
	parts := SliceFlatMap([]string{"alpha bravo", "charlie"}, func(s string) []string {
		return Split(s, " ")
	})
	Println(parts)
	// Output: [alpha bravo charlie]
}

// ExampleSliceTake takes a list prefix through `SliceTake` for agent lists. List
// transforms, predicates, and reducers remain typed core helpers.
func ExampleSliceTake() {
	Println(SliceTake([]string{"alpha", "bravo", "charlie"}, 2))
	// Output: [alpha bravo]
}

// ExampleSliceDrop drops a list prefix through `SliceDrop` for agent lists. List
// transforms, predicates, and reducers remain typed core helpers.
func ExampleSliceDrop() {
	Println(SliceDrop([]string{"alpha", "bravo", "charlie"}, 1))
	// Output: [bravo charlie]
}

// ExampleSliceAny checks whether any list item matches through `SliceAny` for agent lists.
// List transforms, predicates, and reducers remain typed core helpers.
func ExampleSliceAny() {
	Println(SliceAny([]int{1, 2, 3}, func(n int) bool { return n > 2 }))
	// Output: true
}

// ExampleSliceAll checks whether every list item matches through `SliceAll` for agent
// lists. List transforms, predicates, and reducers remain typed core helpers.
func ExampleSliceAll() {
	Println(SliceAll([]int{1, 2, 3}, func(n int) bool { return n > 0 }))
	// Output: true
}
