package core_test

import (
	. "dappco.re/go"
)

// ExampleRegistry_Set sets a value through `Registry.Set` for service registries.
// Registries can list, lock, seal, disable, and reopen named services predictably.
func ExampleRegistry_Set() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Set("bravo", "second")
	Println(r.Get("alpha").Value)
	// Output: first
}

// ExampleNewRegistry_registry constructs a registry through `NewRegistry` for service
// registries. Registries can list, lock, seal, disable, and reopen named services
// predictably.
func ExampleNewRegistry_registry() {
	r := NewRegistry[string]()
	Println(r.Len())
	// Output: 0
}

// ExampleRegistry_Get retrieves a value through `Registry.Get` for service registries.
// Registries can list, lock, seal, disable, and reopen named services predictably.
func ExampleRegistry_Get() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	Println(r.Get("alpha").Value)
	Println(r.Get("missing").OK)
	// Output:
	// first
	// false
}

// ExampleRegistry_Has checks for a value through `Registry.Has` for service registries.
// Registries can list, lock, seal, disable, and reopen named services predictably.
func ExampleRegistry_Has() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	Println(r.Has("alpha"))
	Println(r.Has("missing"))
	// Output:
	// true
	// false
}

// ExampleRegistry_Names lists names through `Registry.Names` for service registries.
// Registries can list, lock, seal, disable, and reopen named services predictably.
func ExampleRegistry_Names() {
	r := NewRegistry[int]()
	r.Set("charlie", 3)
	r.Set("alpha", 1)
	r.Set("bravo", 2)
	Println(r.Names())
	// Output: [charlie alpha bravo]
}

// ExampleRegistry_List lists entries through `Registry.List` for service registries.
// Registries can list, lock, seal, disable, and reopen named services predictably.
func ExampleRegistry_List() {
	r := NewRegistry[string]()
	r.Set("process.run", "run")
	r.Set("process.kill", "kill")
	r.Set("brain.recall", "recall")

	items := r.List("process.*")
	Println(len(items))
	// Output: 2
}

// ExampleRegistry_Each iterates entries through `Registry.Each` for service registries.
// Registries can list, lock, seal, disable, and reopen named services predictably.
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

// ExampleRegistry_Disable hides a registered name from iteration without deleting it.
// Registries can list, lock, seal, disable, and reopen named services
// predictably.
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

// ExampleRegistry_Delete deletes a value through `Registry.Delete` for service registries.
// Registries can list, lock, seal, disable, and reopen named services predictably.
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

// ExampleRegistry_Len counts entries through `Registry.Len` for service registries.
// Registries can list, lock, seal, disable, and reopen named services predictably.
func ExampleRegistry_Len() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Set("bravo", "second")
	Println(r.Len())
	// Output: 2
}

// ExampleRegistry_Enable restores a disabled registry entry to normal iteration.
// Registries can list, lock, seal, disable, and reopen named services
// predictably.
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

// ExampleRegistry_Disabled checks disabled state through `Registry.Disabled` for service
// registries. Registries can list, lock, seal, disable, and reopen named services
// predictably.
func ExampleRegistry_Disabled() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Disable("alpha")
	Println(r.Disabled("alpha"))
	// Output: true
}

// ExampleRegistry_Lock_freeze freezes registry mutation through `Registry.Lock` for
// service registries. Registries can list, lock, seal, disable, and reopen named services
// predictably.
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

// ExampleRegistry_Locked checks locked state through `Registry.Locked` for service
// registries. Registries can list, lock, seal, disable, and reopen named services
// predictably.
func ExampleRegistry_Locked() {
	r := NewRegistry[string]()
	Println(r.Locked())
	r.Lock()
	Println(r.Locked())
	// Output:
	// false
	// true
}

// ExampleRegistry_Seal_shape documents the sealed shape through `Registry.Seal` for
// service registries. Registries can list, lock, seal, disable, and reopen named services
// predictably.
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

// ExampleRegistry_Sealed checks sealed state through `Registry.Sealed` for service
// registries. Registries can list, lock, seal, disable, and reopen named services
// predictably.
func ExampleRegistry_Sealed() {
	r := NewRegistry[string]()
	Println(r.Sealed())
	r.Seal()
	Println(r.Sealed())
	// Output:
	// false
	// true
}

// ExampleRegistry_Open reopens a locked registry so later writes can succeed. Registries
// can list, lock, seal, disable, and reopen named services predictably.
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
