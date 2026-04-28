// SPDX-License-Identifier: EUPL-1.2

// Thread-safe named collection primitive for the Core framework.
// Registry[T] is the universal brick — all named registries (services,
// commands, actions, drives, data) embed this type.
//
// Usage:
//
//	r := core.NewRegistry[*MyService]()
//	r.Set("brain", brainSvc)
//	r.Get("brain")              // Result{brainSvc, true}
//	r.Has("brain")              // true
//	r.Names()                   // []string{"brain"} (insertion order)
//	r.Each(func(name string, svc *MyService) { ... })
//	r.Lock()                    // fully frozen — no more writes
//	r.Seal()                    // no new keys, updates to existing OK
//
// Three lock modes:
//
//	Open   (default) — anything goes
//	Sealed — no new keys, existing keys CAN be updated
//	Locked — fully frozen, no writes at all
package core

import (
	"path/filepath"
	"sync"
)

// registryMode controls write behaviour.
type registryMode int

const (
	registryOpen   registryMode = iota // anything goes
	registrySealed                     // update existing, no new keys
	registryLocked                     // fully frozen
)

// Registry is a thread-safe named collection. The universal brick
// for all named registries in Core.
//
//	r := core.NewRegistry[*Service]()
//	r.Set("brain", svc)
//	if r.Has("brain") { ... }
type Registry[T any] struct {
	items    map[string]T
	disabled map[string]bool
	order    []string // insertion order
	mu       sync.RWMutex
	mode     registryMode
}

// NewRegistry creates an empty registry in Open mode.
//
//	r := core.NewRegistry[*Service]()
func NewRegistry[T any]() *Registry[T] {
	return &Registry[T]{
		items:    make(map[string]T),
		disabled: make(map[string]bool),
	}
}

// Set registers an item by name. Returns Result{OK: false} if the
// registry is locked, or if sealed and the key doesn't already exist.
//
//	r.Set("brain", brainSvc)
func (r *Registry[T]) Set(name string, item T) Result {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch r.mode {
	case registryLocked:
		return Result{E("registry.Set", Concat("registry is locked, cannot set: ", name), nil), false}
	case registrySealed:
		if _, exists := r.items[name]; !exists {
			return Result{E("registry.Set", Concat("registry is sealed, cannot add new key: ", name), nil), false}
		}
	}

	if _, exists := r.items[name]; !exists {
		r.order = append(r.order, name)
	}
	r.items[name] = item
	return Result{OK: true}
}

// Get retrieves an item by name.
//
//	res := r.Get("brain")
//	if res.OK { svc := res.Value.(*Service) }
func (r *Registry[T]) Get(name string) Result {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.items[name]
	if !ok {
		return Result{}
	}
	return Result{item, true}
}

// Has returns true if the name exists in the registry.
//
//	if r.Has("brain") { ... }
func (r *Registry[T]) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.items[name]
	return ok
}

// Names returns all registered names in insertion order.
//
//	names := r.Names() // ["brain", "monitor", "process"]
func (r *Registry[T]) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]string, len(r.order))
	copy(out, r.order)
	return out
}

// List returns items whose names match the glob pattern.
// Uses filepath.Match semantics: "*" matches any sequence, "?" matches one char.
//
//	services := r.List("process.*")
func (r *Registry[T]) List(pattern string) []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []T
	for _, name := range r.order {
		if matched, _ := filepath.Match(pattern, name); matched {
			if !r.disabled[name] {
				result = append(result, r.items[name])
			}
		}
	}
	return result
}

// Each iterates over all items in insertion order, calling fn for each.
// Disabled items are skipped.
//
//	r.Each(func(name string, svc *Service) {
//	    fmt.Println(name, svc)
//	})
func (r *Registry[T]) Each(fn func(string, T)) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, name := range r.order {
		if !r.disabled[name] {
			fn(name, r.items[name])
		}
	}
}

// Len returns the number of registered items (including disabled).
//
//	count := r.Len()
func (r *Registry[T]) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items)
}

// Delete removes an item. Returns Result{OK: false} if locked or not found.
//
//	r.Delete("old-service")
func (r *Registry[T]) Delete(name string) Result {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.mode == registryLocked {
		return Result{E("registry.Delete", Concat("registry is locked, cannot delete: ", name), nil), false}
	}
	if _, exists := r.items[name]; !exists {
		return Result{E("registry.Delete", Concat("not found: ", name), nil), false}
	}

	delete(r.items, name)
	delete(r.disabled, name)
	// Remove from order slice
	for i, n := range r.order {
		if n == name {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
	return Result{OK: true}
}

// Disable soft-disables an item. It still exists but Each/List skip it.
// Returns Result{OK: false} if not found.
//
//	r.Disable("broken-handler")
func (r *Registry[T]) Disable(name string) Result {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[name]; !exists {
		return Result{E("registry.Disable", Concat("not found: ", name), nil), false}
	}
	r.disabled[name] = true
	return Result{OK: true}
}

// Enable re-enables a disabled item.
//
//	r.Enable("fixed-handler")
func (r *Registry[T]) Enable(name string) Result {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[name]; !exists {
		return Result{E("registry.Enable", Concat("not found: ", name), nil), false}
	}
	delete(r.disabled, name)
	return Result{OK: true}
}

// Disabled returns true if the item is soft-disabled.
//
//	r := core.NewRegistry[string]()
//	r.Set("agent", "codex")
//	r.Disable("agent")
//	if r.Disabled("agent") { core.Println("disabled") }
func (r *Registry[T]) Disabled(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.disabled[name]
}

// Lock fully freezes the registry. No Set, no Delete.
//
//	r.Lock() // after startup, prevent late registration
func (r *Registry[T]) Lock() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mode = registryLocked
}

// Locked returns true if the registry is fully frozen.
//
//	r := core.NewRegistry[string]()
//	r.Lock()
//	if r.Locked() { core.Println("locked") }
func (r *Registry[T]) Locked() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mode == registryLocked
}

// Seal prevents new keys but allows updates to existing keys.
// Use for hot-reload: shape is fixed, implementations can change.
//
//	r.Seal() // no new capabilities, but handlers can be swapped
func (r *Registry[T]) Seal() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mode = registrySealed
}

// Sealed returns true if the registry is sealed (no new keys).
//
//	r := core.NewRegistry[string]()
//	r.Set("agent", "codex")
//	r.Seal()
//	if r.Sealed() { core.Println("sealed") }
func (r *Registry[T]) Sealed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mode == registrySealed
}

// Open resets the registry to open mode (default).
//
//	r.Open() // re-enable writes for testing
func (r *Registry[T]) Open() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mode = registryOpen
}
