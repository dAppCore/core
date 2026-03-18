// SPDX-License-Identifier: EUPL-1.2

// Package core is a dependency injection and service lifecycle framework for Go.
// This file defines the Core struct, accessors, and IPC/error wrappers.

package core

import (
	"sync"
	"sync/atomic"
)

// --- Core Struct ---

// Core is the central application object that manages services, assets, and communication.
type Core struct {
	app  *App     // c.App()     — Application identity + optional GUI runtime
	emb  *Embed    // c.Embed()   — Mounted embedded assets (read-only)
	fs   *Fs      // c.Fs()      — Local filesystem I/O (sandboxable)
	cfg  *Config    // c.Config()  — Configuration, settings, feature flags
	err  *ErrPan  // c.Error()   — Panic recovery and crash reporting
	log  *ErrLog  // c.Log()     — Structured logging + error wrapping
	cli  *Cli     // c.Cli()     — CLI command framework
	srv  *Service // c.Service("name") — Service registry and lifecycle
	lock *Lock    // c.Lock("name") — Named mutexes
	ipc  *Ipc     // c.IPC()     — Message bus for IPC
	i18n *I18n    // c.I18n()    — Internationalisation and locale collection

	taskIDCounter atomic.Uint64
	wg            sync.WaitGroup
	shutdown      atomic.Bool
}

// --- Accessors ---

func (c *Core) App() *App       { return c.app }
func (c *Core) Embed() *Embed    { return c.emb }
func (c *Core) Fs() *Fs         { return c.fs }
func (c *Core) Config() *Config   { return c.cfg }
func (c *Core) Error() *ErrPan  { return c.err }
func (c *Core) Log() *ErrLog    { return c.log }
func (c *Core) Cli() *Cli       { return c.cli }
func (c *Core) IPC() *Ipc       { return c.ipc }
func (c *Core) I18n() *I18n     { return c.i18n }
func (c *Core) Core() *Core     { return c }

// --- IPC ---

func (c *Core) ACTION(msg Message) error                                { return c.ipc.Action(msg) }
func (c *Core) RegisterAction(handler func(*Core, Message) error)       { c.ipc.RegisterAction(handler) }
func (c *Core) RegisterActions(handlers ...func(*Core, Message) error)  { c.ipc.RegisterActions(handlers...) }
func (c *Core) QUERY(q Query) (any, bool, error)                        { return c.ipc.Query(q) }
func (c *Core) QUERYALL(q Query) ([]any, error)                         { return c.ipc.QueryAll(q) }
func (c *Core) PERFORM(t Task) (any, bool, error)                       { return c.ipc.Perform(t) }
func (c *Core) RegisterQuery(handler QueryHandler)                      { c.ipc.RegisterQuery(handler) }
func (c *Core) RegisterTask(handler TaskHandler)                        { c.ipc.RegisterTask(handler) }

// --- Error+Log ---

// LogError logs an error and returns a wrapped error.
func (c *Core) LogError(err error, op, msg string) error {
	return c.log.Error(err, op, msg)
}

// LogWarn logs a warning and returns a wrapped error.
func (c *Core) LogWarn(err error, op, msg string) error {
	return c.log.Warn(err, op, msg)
}

// Must logs and panics if err is not nil.
func (c *Core) Must(err error, op, msg string) {
	c.log.Must(err, op, msg)
}

// --- Global Instance ---

