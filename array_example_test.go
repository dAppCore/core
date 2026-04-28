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

func ExampleArray_AddUnique() {
	a := NewArray[string]()
	a.AddUnique("alpha")
	a.AddUnique("alpha") // no duplicate
	a.AddUnique("bravo")

	Println(a.Len())
	// Output: 2
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
