package core_test

import (
	. "dappco.re/go"
)

func ExampleRegistry_Set() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Set("bravo", "second")
	Println(r.Get("alpha").Value)
	// Output: first
}

func ExampleNewRegistry_registry() {
	r := NewRegistry[string]()
	Println(r.Len())
	// Output: 0
}

func ExampleRegistry_Get() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	Println(r.Get("alpha").Value)
	Println(r.Get("missing").OK)
	// Output:
	// first
	// false
}

func ExampleRegistry_Has() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	Println(r.Has("alpha"))
	Println(r.Has("missing"))
	// Output:
	// true
	// false
}

func ExampleRegistry_Names() {
	r := NewRegistry[int]()
	r.Set("charlie", 3)
	r.Set("alpha", 1)
	r.Set("bravo", 2)
	Println(r.Names())
	// Output: [charlie alpha bravo]
}

func ExampleRegistry_List() {
	r := NewRegistry[string]()
	r.Set("process.run", "run")
	r.Set("process.kill", "kill")
	r.Set("brain.recall", "recall")

	items := r.List("process.*")
	Println(len(items))
	// Output: 2
}

func ExampleRegistry_Each() {
	r := NewRegistry[int]()
	r.Set("a", 1)
	r.Set("b", 2)
	r.Set("c", 3)

	sum := 0
	r.Each(func(_ string, v int) { sum += v })
	Println(sum)
	// Output: 6
}

func ExampleRegistry_Disable() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Set("bravo", "second")
	r.Disable("alpha")

	var names []string
	r.Each(func(name string, _ string) { names = append(names, name) })
	Println(names)
	// Output: [bravo]
}

func ExampleRegistry_Delete() {
	r := NewRegistry[string]()
	r.Set("temp", "value")
	Println(r.Has("temp"))

	r.Delete("temp")
	Println(r.Has("temp"))
	// Output:
	// true
	// false
}

func ExampleRegistry_Len() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Set("bravo", "second")
	Println(r.Len())
	// Output: 2
}

func ExampleRegistry_Enable() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Disable("alpha")
	Println(r.Disabled("alpha"))
	r.Enable("alpha")
	Println(r.Disabled("alpha"))
	// Output:
	// true
	// false
}

func ExampleRegistry_Disabled() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Disable("alpha")
	Println(r.Disabled("alpha"))
	// Output: true
}

func ExampleRegistry_Lock_freeze() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Lock()
	Println(r.Locked())
	Println(r.Set("bravo", "second").OK)
	// Output:
	// true
	// false
}

func ExampleRegistry_Locked() {
	r := NewRegistry[string]()
	Println(r.Locked())
	r.Lock()
	Println(r.Locked())
	// Output:
	// false
	// true
}

func ExampleRegistry_Seal_shape() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Seal()
	Println(r.Sealed())
	Println(r.Set("alpha", "updated").OK)
	Println(r.Set("bravo", "second").OK)
	// Output:
	// true
	// true
	// false
}

func ExampleRegistry_Sealed() {
	r := NewRegistry[string]()
	Println(r.Sealed())
	r.Seal()
	Println(r.Sealed())
	// Output:
	// false
	// true
}

func ExampleRegistry_Open() {
	r := NewRegistry[string]()
	r.Lock()
	r.Open()
	Println(r.Locked())
	Println(r.Set("alpha", "first").OK)
	// Output:
	// false
	// true
}
