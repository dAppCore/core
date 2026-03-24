// SPDX-License-Identifier: EUPL-1.2

// Package core is a dependency injection and service lifecycle framework for Go.
// This file defines the Core struct, accessors, and IPC/error wrappers.

package core

import (
	"context"
	"sync"
	"sync/atomic"
)

// --- Core Struct ---

// Core is the central application object that manages services, assets, and communication.
type Core struct {
	options *Options    // c.Options()        — Input configuration used to create this Core
	app     *App        // c.App()            — Application identity + optional GUI runtime
	data    *Data       // c.Data()           — Embedded/stored content from packages
	drive   *Drive      // c.Drive()          — Resource handle registry (transports)
	fs      *Fs         // c.Fs()             — Local filesystem I/O (sandboxable)
	config  *Config     // c.Config()         — Configuration, settings, feature flags
	error   *ErrorPanic // c.Error()          — Panic recovery and crash reporting
	log     *ErrorLog   // c.Log()            — Structured logging + error wrapping
	// cli accessed via ServiceFor[*Cli](c, "cli")
	commands *commandRegistry // c.Command("path")  — Command tree
	services *serviceRegistry // c.Service("name")  — Service registry
	lock     *Lock            // c.Lock("name")     — Named mutexes
	ipc      *Ipc             // c.IPC()            — Message bus for IPC
	info     *SysInfo         // c.Env("key")        — Read-only system/environment information
	i18n     *I18n            // c.I18n()           — Internationalisation and locale collection

	context       context.Context
	cancel        context.CancelFunc
	taskIDCounter atomic.Uint64
	waitGroup     sync.WaitGroup
	shutdown      atomic.Bool
}

// --- Accessors ---

func (c *Core) Options() *Options        { return c.options }
func (c *Core) App() *App                { return c.app }
func (c *Core) Data() *Data              { return c.data }
func (c *Core) Drive() *Drive            { return c.drive }
func (c *Core) Embed() Result            { return c.data.Get("app") } // legacy — use Data()
func (c *Core) Fs() *Fs                  { return c.fs }
func (c *Core) Config() *Config          { return c.config }
func (c *Core) Error() *ErrorPanic       { return c.error }
func (c *Core) Log() *ErrorLog           { return c.log }
func (c *Core) Cli() *Cli                { return cli.New() }
func (c *Core) IPC() *Ipc                { return c.ipc }
func (c *Core) I18n() *I18n              { return c.i18n }
func (c *Core) Env(key string) string    { return Env(key) }
func (c *Core) Context() context.Context { return c.context }
func (c *Core) Core() *Core              { return c }

// --- IPC (uppercase aliases) ---

func (c *Core) ACTION(msg Message) Result { return c.Action(msg) }
func (c *Core) QUERY(q Query) Result      { return c.Query(q) }
func (c *Core) QUERYALL(q Query) Result   { return c.QueryAll(q) }
func (c *Core) PERFORM(t Task) Result     { return c.Perform(t) }

// --- Error+Log ---

// LogError logs an error and returns the Result from ErrorLog.
func (c *Core) LogError(err error, op, msg string) Result {
	return c.log.Error(err, op, msg)
}

// LogWarn logs a warning and returns the Result from ErrorLog.
func (c *Core) LogWarn(err error, op, msg string) Result {
	return c.log.Warn(err, op, msg)
}

// Must logs and panics if err is not nil.
func (c *Core) Must(err error, op, msg string) {
	c.log.Must(err, op, msg)
}

// --- Global Instance ---
