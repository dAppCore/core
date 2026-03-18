// SPDX-License-Identifier: EUPL-1.2

// Package core is the Core framework for Go.
//
// Single import, single struct, everything accessible:
//
//	import core "forge.lthn.ai/core/go"
//
//	c, _ := core.New(
//	    core.WithAssets(myEmbed),
//	    core.WithService(myFactory),
//	)
//
//	// DI
//	svc, _ := core.ServiceFor[*MyService](c, "name")
//
//	// Mount
//	content, _ := c.Mnt().ReadString("persona/secops/developer.md")
//	c.Mnt().Extract(targetDir, data)
//
//	// IPC
//	c.ACTION(msg)
package core

import (
	di "forge.lthn.ai/core/go/pkg/core"
)

// --- Types ---

// Core is the central application container.
type Core = di.Core

// Option configures a Core instance.
type Option = di.Option

// Message is the IPC message type.
type Message = di.Message

// Sub is a scoped view of an embedded filesystem.
type Sub = di.Sub

// ExtractOptions configures template extraction.
type ExtractOptions = di.ExtractOptions

// Startable is implemented by services with startup logic.
type Startable = di.Startable

// Stoppable is implemented by services with shutdown logic.
type Stoppable = di.Stoppable

// LocaleProvider provides locale filesystems for i18n.
type LocaleProvider = di.LocaleProvider

// ServiceRuntime is the base for services with typed options.
type ServiceRuntime[T any] = di.ServiceRuntime[T]

// --- Constructor + Options ---

// New creates a new Core instance.
var New = di.New

// WithService registers a service factory.
var WithService = di.WithService

// WithName registers a named service factory.
var WithName = di.WithName

// WithAssets mounts an embedded filesystem.
var WithAssets = di.WithAssets

// WithMount mounts an embedded filesystem at a subdirectory.
var WithMount = di.WithMount

// WithServiceLock prevents late service registration.
var WithServiceLock = di.WithServiceLock

// WithApp sets the GUI runtime.
var WithApp = di.WithApp

// Mount creates a scoped view of an embed.FS at basedir.
var Mount = di.Mount

// --- Generic Functions ---

// ServiceFor retrieves a typed service by name.
func ServiceFor[T any](c *Core, name string) (T, error) {
	return di.ServiceFor[T](c, name)
}

// E creates a structured error.
var E = di.E

// --- Configuration (core.Etc) ---

// Etc is the configuration and feature flags store.
type Etc = di.Etc

// NewEtc creates a standalone configuration store.
var NewEtc = di.NewEtc

// Var is a typed optional variable (set/unset/get).
type Var[T any] = di.Var[T]

// NewVar creates a Var with the given value.
func NewVar[T any](val T) Var[T] {
	return di.NewVar(val)
}
