// SPDX-License-Identifier: EUPL-1.2

// Package core is the Core framework for Go.
//
// Import this package to access the full framework surface:
//
//	import core "forge.lthn.ai/core/go"
//
//	c, _ := core.New(core.WithService(myFactory))
//	svc, _ := core.ServiceFor[*MyService](c, "name")
//
// Sub-packages provide domain-specific capabilities:
//
//	core/pkg/core  — DI container, ServiceRuntime, lifecycle
//	core/pkg/mnt   — mount operations (embed FS, template extraction)
//	core/pkg/log   — structured logging (re-exported from go-log)
//
// The framework object is designed for zero transitive dependencies.
// Each pkg/ uses stdlib only.
package core

import (
	"forge.lthn.ai/core/go/pkg/core"
)

// Re-export the DI container API at the top level.
// This lets users write core.New() instead of core.Core.New().

// Core is the central application container.
type Core = core.Core

// New creates a new Core instance with the given options.
var New = core.New

// Option configures a Core instance.
type Option = core.Option

// WithService registers a service factory.
var WithService = core.WithService

// WithName registers a named service factory.
var WithName = core.WithName

// WithServiceLock prevents late service registration.
var WithServiceLock = core.WithServiceLock

// WithAssets registers an embedded filesystem.
var WithAssets = core.WithAssets

// ServiceFor retrieves a typed service by name.
func ServiceFor[T any](c *Core, name string) (T, error) {
	return core.ServiceFor[T](c, name)
}

// ServiceRuntime is the base for services with typed options.
type ServiceRuntime[T any] = core.ServiceRuntime[T]

// NewServiceRuntime creates a ServiceRuntime — use pkg/core.NewServiceRuntime[T] directly.
// Cannot re-export generic functions at the package level.

// Message is the IPC message type.
type Message = core.Message

// Startable is implemented by services with startup logic.
type Startable = core.Startable

// Stoppable is implemented by services with shutdown logic.
type Stoppable = core.Stoppable

// LocaleProvider is implemented by services that provide locale filesystems.
type LocaleProvider = core.LocaleProvider
