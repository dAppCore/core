package core_test

import (
	. "dappco.re/go"
)

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

func ExampleArray_Add() {
	a := NewArray[string]()
	a.Add("alpha", "bravo")
	Println(a.AsSlice())
	// Output: [alpha bravo]
}

func ExampleArray_AddUnique() {
	a := NewArray[string]()
	a.AddUnique("alpha")
	a.AddUnique("alpha") // no duplicate
	a.AddUnique("bravo")

	Println(a.Len())
	// Output: 2
}

func ExampleArray_Contains() {
	a := NewArray("alpha", "bravo")
	Println(a.Contains("bravo"))
	Println(a.Contains("charlie"))
	// Output:
	// true
	// false
}

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

func ExampleArray_Each() {
	a := NewArray("alpha", "bravo", "charlie")
	var labels []string
	a.Each(func(item string) {
		labels = append(labels, Upper(item[:1]))
	})
	Println(labels)
	// Output: [A B C]
}

func ExampleArray_Remove() {
	a := NewArray("alpha", "bravo", "charlie")
	a.Remove("bravo")
	Println(a.AsSlice())
	// Output: [alpha charlie]
}

func ExampleArray_Deduplicate() {
	a := NewArray("alpha", "alpha", "bravo")
	a.Deduplicate()
	Println(a.AsSlice())
	// Output: [alpha bravo]
}

func ExampleArray_Len() {
	a := NewArray("alpha", "bravo")
	Println(a.Len())
	// Output: 2
}

func ExampleArray_Clear() {
	a := NewArray("alpha", "bravo")
	a.Clear()
	Println(a.Len())
	// Output: 0
}

func ExampleArray_AsSlice() {
	a := NewArray("alpha", "bravo")
	Println(a.AsSlice())
	// Output: [alpha bravo]
}
