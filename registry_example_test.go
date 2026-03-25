package core_test

import (
	"fmt"

	. "dappco.re/go/core"
)

func ExampleRegistry_Set() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Set("bravo", "second")
	fmt.Println(r.Get("alpha").Value)
	// Output: first
}

func ExampleRegistry_Names() {
	r := NewRegistry[int]()
	r.Set("charlie", 3)
	r.Set("alpha", 1)
	r.Set("bravo", 2)
	fmt.Println(r.Names())
	// Output: [charlie alpha bravo]
}

func ExampleRegistry_List() {
	r := NewRegistry[string]()
	r.Set("process.run", "run")
	r.Set("process.kill", "kill")
	r.Set("brain.recall", "recall")

	items := r.List("process.*")
	fmt.Println(len(items))
	// Output: 2
}

func ExampleRegistry_Each() {
	r := NewRegistry[int]()
	r.Set("a", 1)
	r.Set("b", 2)
	r.Set("c", 3)

	sum := 0
	r.Each(func(_ string, v int) { sum += v })
	fmt.Println(sum)
	// Output: 6
}

func ExampleRegistry_Disable() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Set("bravo", "second")
	r.Disable("alpha")

	var names []string
	r.Each(func(name string, _ string) { names = append(names, name) })
	fmt.Println(names)
	// Output: [bravo]
}

func ExampleRegistry_Delete() {
	r := NewRegistry[string]()
	r.Set("temp", "value")
	fmt.Println(r.Has("temp"))

	r.Delete("temp")
	fmt.Println(r.Has("temp"))
	// Output:
	// true
	// false
}
