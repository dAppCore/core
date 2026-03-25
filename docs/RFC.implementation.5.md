# Implementation Plan 5 — Missing Primitives + AX-7 (Phase 1-2)

> Independent of Plans 2-4. Can ship alongside or before.

## core.ID() — Unique Identifier

**File:** `utils.go` or new `id.go`

```go
var idCounter atomic.Uint64

func ID() string {
    return Sprintf("id-%d-%s", idCounter.Add(1), shortRand())
}

func shortRand() string {
    b := make([]byte, 3)
    crand.Read(b)
    return hex.EncodeToString(b)
}
```

Replaces 3 different patterns: crypto/rand hex, atomic counter, fmt.Sprintf.

## core.ValidateName() / core.ValidatePath()

**File:** `utils.go` or new `validate.go`

```go
func ValidateName(name string) Result {
    if name == "" || name == "." || name == ".." {
        return Result{E("validate", "invalid name: "+name, nil), false}
    }
    if Contains(name, "/") || Contains(name, "\\") {
        return Result{E("validate", "name contains path separator: "+name, nil), false}
    }
    return Result{name, true}
}

func SanitisePath(path string) string {
    safe := PathBase(path)
    if safe == "." || safe == ".." || safe == "" {
        return "invalid"
    }
    return safe
}
```

Replaces copy-pasted validation in prep.go, plan.go, command.go.

## Fs.WriteAtomic()

See Plan 4 Phase E. Fixes P4-9 and P4-10.

## Fs.NewUnrestricted()

**File:** `fs.go`

```go
func (m *Fs) NewUnrestricted() *Fs {
    return m.New("/")
}
```

Gives consumers a legitimate door instead of unsafe.Pointer crowbar.
Fixes P11-2. Add go vet / linter rule to flag unsafe.Pointer on Core types.

## AX-7 for core/go

Currently 14% AX-7 (83.6% statement coverage, wrong naming).

Steps:
1. Run the rename script from core/agent (same Python script)
2. Gap analysis: find functions missing Good/Bad/Ugly
3. Fill gaps — 212 functions × 3 categories = 636 target
4. Currently 91/636 filled → need 545 more

Prioritise by section:
- Section 3 (services): ServiceFor, RegisterService, Service
- Section 4 (IPC): Action, Query, RegisterAction
- Section 8 (Fs): Read, Write, WriteAtomic, Delete
- Section 5 (Config): Get, Set, Enable, ConfigGet

## RunE() — Backwards-Compatible Run Fix

**File:** `core.go`

```go
func (c *Core) RunE() error {
    defer c.ServiceShutdown(context.Background())

    r := c.ServiceStartup(c.context, nil)
    if !r.OK {
        if err, ok := r.Value.(error); ok {
            return err
        }
        return E("core.Run", "startup failed", nil)
    }

    if cli := c.Cli(); cli != nil {
        r = cli.Run()
    }

    if !r.OK {
        if err, ok := r.Value.(error); ok {
            return err
        }
    }
    return nil
}

// Run keeps backwards compatibility
func (c *Core) Run() {
    if err := c.RunE(); err != nil {
        Error(err.Error())
        os.Exit(1)
    }
}
```

Fixes P7-5 (os.Exit bypasses defer) without breaking 15 main.go files.
