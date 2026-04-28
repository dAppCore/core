package core_test

import (
	. "dappco.re/go"
)

// ExampleNewArray creates a typed array through `NewArray` for an in-memory agent queue.
// Typed collection helpers cover queue operations without exposing slices directly.
func ExampleNewArray() {
	a := NewArray[string]()
	a.Add("alpha")
	a.Add("bravo")
	a.Add("charlie")

	Println(a.Len())
	Println(a.Contains("bravo"))
	// Output:
	// 3
	// true
}

// ExampleArray_Add adds a value through `Array.Add` for an in-memory agent queue. Typed
// collection helpers cover queue operations without exposing slices directly.
func ExampleArray_Add() {
	a := NewArray[string]()
	a.Add("alpha", "bravo")
	Println(a.AsSlice())
	// Output: [alpha bravo]
}

// ExampleArray_AddUnique adds a value through `Array.AddUnique` only when it is absent for
// an in-memory agent queue. Typed collection helpers cover queue operations without
// exposing slices directly.
func ExampleArray_AddUnique() {
	a := NewArray[string]()
	a.AddUnique("alpha")
	a.AddUnique("alpha") // no duplicate
	a.AddUnique("bravo")

	Println(a.Len())
	// Output: 2
}

// ExampleArray_Contains checks text membership through `Array.Contains` for an in-memory
// agent queue. Typed collection helpers cover queue operations without exposing slices
// directly.
func ExampleArray_Contains() {
	a := NewArray("alpha", "bravo")
	Println(a.Contains("bravo"))
	Println(a.Contains("charlie"))
	// Output:
	// true
	// false
}

// ExampleArray_Filter filters values through `Array.Filter` for an in-memory agent queue.
// Typed collection helpers cover queue operations without exposing slices directly.
func ExampleArray_Filter() {
	a := NewArray[int]()
	a.Add(1)
	a.Add(2)
	a.Add(3)
	a.Add(4)

	r := a.Filter(func(n int) bool { return n%2 == 0 })
	Println(r.OK)
	// Output: true
}

// ExampleArray_Each iterates entries through `Array.Each` for an in-memory agent queue.
// Typed collection helpers cover queue operations without exposing slices directly.
func ExampleArray_Each() {
	a := NewArray("alpha", "bravo", "charlie")
	var labels []string
	a.Each(func(item string) {
		labels = append(labels, Upper(item[:1]))
	})
	Println(labels)
	// Output: [A B C]
}

// ExampleArray_Remove removes a value through `Array.Remove` for an in-memory agent queue.
// Typed collection helpers cover queue operations without exposing slices directly.
func ExampleArray_Remove() {
	a := NewArray("alpha", "bravo", "charlie")
	a.Remove("bravo")
	Println(a.AsSlice())
	// Output: [alpha charlie]
}

// ExampleArray_Deduplicate deduplicates values through `Array.Deduplicate` for an
// in-memory agent queue. Typed collection helpers cover queue operations without exposing
// slices directly.
func ExampleArray_Deduplicate() {
	a := NewArray("alpha", "alpha", "bravo")
	a.Deduplicate()
	Println(a.AsSlice())
	// Output: [alpha bravo]
}

// ExampleArray_Len counts entries through `Array.Len` for an in-memory agent queue. Typed
// collection helpers cover queue operations without exposing slices directly.
func ExampleArray_Len() {
	a := NewArray("alpha", "bravo")
	Println(a.Len())
	// Output: 2
}

// ExampleArray_Clear clears entries through `Array.Clear` for an in-memory agent queue.
// Typed collection helpers cover queue operations without exposing slices directly.
func ExampleArray_Clear() {
	a := NewArray("alpha", "bravo")
	a.Clear()
	Println(a.Len())
	// Output: 0
}

// ExampleArray_AsSlice exports values through `Array.AsSlice` as a slice for an in-memory
// agent queue. Typed collection helpers cover queue operations without exposing slices
// directly.
func ExampleArray_AsSlice() {
	a := NewArray("alpha", "bravo")
	Println(a.AsSlice())
	// Output: [alpha bravo]
}
