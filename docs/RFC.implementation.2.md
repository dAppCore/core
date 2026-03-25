# Implementation Plan 2 — Registry[T] Primitive (Section 20)

> Foundation brick. Most other plans depend on this.

## The Type

**New file:** `registry.go`

```go
type Registry[T any] struct {
    items  map[string]T
    order  []string        // insertion order (fixes P4-1)
    mu     sync.RWMutex
    mode   registryMode    // open, sealed, locked
}

type registryMode int
const (
    registryOpen   registryMode = iota  // anything goes
    registrySealed                       // update existing, no new keys
    registryLocked                       // fully frozen
)
```

## Methods

```go
func NewRegistry[T any]() *Registry[T]
func (r *Registry[T]) Set(name string, item T) Result
func (r *Registry[T]) Get(name string) Result
func (r *Registry[T]) Has(name string) bool
func (r *Registry[T]) Names() []string          // insertion order
func (r *Registry[T]) List(pattern string) []T  // glob match
func (r *Registry[T]) Each(fn func(string, T))
func (r *Registry[T]) Len() int
func (r *Registry[T]) Delete(name string) Result
func (r *Registry[T]) Disable(name string)      // soft — exists but skipped
func (r *Registry[T]) Lock()                     // fully frozen
func (r *Registry[T]) Seal()                     // no new, updates OK
func (r *Registry[T]) Open()                     // default
```

## Migration

Replace internal registries one at a time:

1. `serviceRegistry` → `ServiceRegistry` embedding `Registry[*Service]`
2. `commandRegistry` → `CommandRegistry` embedding `Registry[*Command]`
3. `Drive.handles` → embed `Registry[*DriveHandle]`
4. `Data.mounts` → embed `Registry[*Embed]`
5. `Lock.locks` → `Registry[*sync.RWMutex]` (fixes P4-8 allocation)

Each migration is a separate commit. Tests before and after.

## Core Accessor

```go
func (c *Core) Registry(name string) *Registry[any]
```

Returns named registries for cross-cutting queries.

## Resolves

P4-1 (startup order), P4-8 (lock allocation), I6 (serviceRegistry unexported), I12 (Ipc data-only), I13 (Lock allocation), I14 (Startables return type).
