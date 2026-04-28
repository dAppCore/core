package core_test

import . "dappco.re/go"

func ExampleSliceContains() {
	Println(SliceContains([]string{"alpha", "bravo"}, "bravo"))
	// Output: true
}

func ExampleSliceIndex() {
	Println(SliceIndex([]string{"alpha", "bravo"}, "bravo"))
	// Output: 1
}

func ExampleSliceSort() {
	names := []string{"charlie", "alpha", "bravo"}
	SliceSort(names)
	Println(names)
	// Output: [alpha bravo charlie]
}

func ExampleSliceUniq() {
	Println(SliceUniq([]string{"alpha", "alpha", "bravo"}))
	// Output: [alpha bravo]
}

func ExampleSliceReverse() {
	names := []string{"alpha", "bravo", "charlie"}
	SliceReverse(names)
	Println(names)
	// Output: [charlie bravo alpha]
}

func ExampleSliceFilter() {
	evens := SliceFilter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
	Println(evens)
	// Output: [2 4]
}

func ExampleSliceMap() {
	labels := SliceMap([]int{1, 2, 3}, func(n int) string { return Sprintf("node-%d", n) })
	Println(labels)
	// Output: [node-1 node-2 node-3]
}

func ExampleSliceReduce() {
	sum := SliceReduce([]int{1, 2, 3}, 0, func(acc, n int) int { return acc + n })
	Println(sum)
	// Output: 6
}

func ExampleSliceFlatMap() {
	parts := SliceFlatMap([]string{"alpha bravo", "charlie"}, func(s string) []string {
		return Split(s, " ")
	})
	Println(parts)
	// Output: [alpha bravo charlie]
}

func ExampleSliceTake() {
	Println(SliceTake([]string{"alpha", "bravo", "charlie"}, 2))
	// Output: [alpha bravo]
}

func ExampleSliceDrop() {
	Println(SliceDrop([]string{"alpha", "bravo", "charlie"}, 1))
	// Output: [bravo charlie]
}

func ExampleSliceAny() {
	Println(SliceAny([]int{1, 2, 3}, func(n int) bool { return n > 2 }))
	// Output: true
}

func ExampleSliceAll() {
	Println(SliceAll([]int{1, 2, 3}, func(n int) bool { return n > 0 }))
	// Output: true
}
