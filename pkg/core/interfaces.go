package core

import (
	"context"
	goio "io"
	"io/fs"
	"sync"
	"sync/atomic"
)

// This file defines the public API contracts (interfaces) for the services
// in the Core framework. Services depend on these interfaces, not on
// concrete implementations.

// Contract specifies the operational guarantees that the Core and its services must adhere to.
// This is used for configuring panic handling and other resilience features.
type Contract struct {
	// DontPanic, if true, instructs the Core to recover from panics and return an error instead.
	DontPanic bool
	// DisableLogging, if true, disables all logging from the Core and its services.
	DisableLogging bool
}

// Option is a function that configures the Core.
// This is used to apply settings and register services during initialization.
type Option func(*Core) error

// Message is the interface for all messages that can be sent through the Core's IPC system.
// Any struct can be a message, allowing for structured data to be passed between services.
// Used with ACTION for fire-and-forget broadcasts.
type Message any

// Query is the interface for read-only requests that return data.
// Used with QUERY (first responder) or QUERYALL (all responders).
type Query any

// Task is the interface for requests that perform side effects.
// Used with PERFORM (first responder executes).
type Task any

// TaskWithID is an optional interface for tasks that need to know their assigned ID.
// This is useful for tasks that want to report progress back to the frontend.
type TaskWithID interface {
	Task
	SetTaskID(id string)
	GetTaskID() string
}

// QueryHandler handles Query requests. Returns (result, handled, error).
// If handled is false, the query will be passed to the next handler.
type QueryHandler func(*Core, Query) (any, bool, error)

// TaskHandler handles Task requests. Returns (result, handled, error).
// If handled is false, the task will be passed to the next handler.
type TaskHandler func(*Core, Task) (any, bool, error)

// Startable is an interface for services that need to perform initialization.
type Startable interface {
	OnStartup(ctx context.Context) error
}

// Stoppable is an interface for services that need to perform cleanup.
type Stoppable interface {
	OnShutdown(ctx context.Context) error
}

// LocaleProvider is implemented by services that ship their own translation files.
// Core discovers this interface during service registration and collects the
// locale filesystems. The i18n service loads them during startup.
//
// Usage in a service package:
//
//	//go:embed locales
//	var localeFS embed.FS
//
//	func (s *MyService) Locales() fs.FS { return localeFS }
type LocaleProvider interface {
	Locales() fs.FS
}

// Core is the central application object that manages services, assets, and communication.
type Core struct {
	App   any           // GUI runtime (e.g., Wails App) - set by WithApp option
	mnt   *Sub          // Mounted embedded assets (read-only)
	io    *IO           // Local filesystem I/O (read/write, sandboxable)
	etc   *Etc          // Configuration, settings, and feature flags
	crash *CrashHandler // Panic recovery and crash reporting
	cli   *CliApp       // CLI command registration and execution
	svc      *serviceManager
	bus      *messageBus
	locales  []fs.FS // collected from LocaleProvider services

	taskIDCounter atomic.Uint64
	wg            sync.WaitGroup
	shutdown      atomic.Bool
}

// Mnt returns the mounted embedded assets (read-only).
//
//	c.Mnt().ReadString("persona/secops/developer.md")
func (c *Core) Mnt() *Sub {
	return c.mnt
}

// Io returns the local filesystem I/O layer.
// Default: rooted at "/". Sandboxable via WithIO("./data").
//
//	c.Io().Read("config.yaml")
//	c.Io().Write("output.txt", content)
func (c *Core) Io() *IO {
	return c.io
}

// Etc returns the configuration and feature flags store.
//
//	c.Etc().Set("api_url", "https://api.lthn.sh")
//	c.Etc().Enable("coderabbit")
//	c.Etc().Enabled("coderabbit") // true
func (c *Core) Etc() *Etc {
	return c.etc
}

// Crash returns the crash handler for panic recovery.
//
//	defer c.Crash().Recover()
//	c.Crash().SafeGo(func() { ... })
func (c *Core) Crash() *CrashHandler {
	return c.crash
}

// Cli returns the CLI command framework.
// Register commands without importing any CLI package.
//
//	c.Cli().NewSubCommand("health", "Check service health").Action(func() error { ... })
func (c *Core) Cli() *CliApp {
	return c.cli
}

// Locales returns all locale filesystems collected from registered services.
// The i18n service uses this during startup to load translations.
func (c *Core) Locales() []fs.FS {
	return c.locales
}

// Config provides access to application configuration.
type Config interface {
	// Get retrieves a configuration value by key and stores it in the 'out' variable.
	Get(key string, out any) error
	// Set stores a configuration value by key.
	Set(key string, v any) error
}

// WindowOption is an interface for applying configuration options to a window.
type WindowOption interface {
	Apply(any)
}

// Display provides access to windowing and visual elements.
type Display interface {
	// OpenWindow creates a new window with the given options.
	OpenWindow(opts ...WindowOption) error
}

// Workspace provides management for encrypted user workspaces.
type Workspace interface {
	// CreateWorkspace creates a new encrypted workspace.
	CreateWorkspace(identifier, password string) (string, error)
	// SwitchWorkspace changes the active workspace.
	SwitchWorkspace(name string) error
	// WorkspaceFileGet retrieves the content of a file from the active workspace.
	WorkspaceFileGet(filename string) (string, error)
	// WorkspaceFileSet saves content to a file in the active workspace.
	WorkspaceFileSet(filename, content string) error
}

// Crypt provides PGP-based encryption, signing, and key management.
type Crypt interface {
	// CreateKeyPair generates a new PGP keypair.
	CreateKeyPair(name, passphrase string) (string, error)
	// EncryptPGP encrypts data for a recipient.
	EncryptPGP(writer goio.Writer, recipientPath, data string, opts ...any) (string, error)
	// DecryptPGP decrypts a PGP message.
	DecryptPGP(recipientPath, message, passphrase string, opts ...any) (string, error)
}

// ActionServiceStartup is a message sent when the application's services are starting up.
// This provides a hook for services to perform initialization tasks.
type ActionServiceStartup struct{}

// ActionServiceShutdown is a message sent when the application is shutting down.
// This allows services to perform cleanup tasks, such as saving state or closing resources.
type ActionServiceShutdown struct{}

// ActionTaskStarted is a message sent when a background task has started.
type ActionTaskStarted struct {
	TaskID string
	Task   Task
}

// ActionTaskProgress is a message sent when a task has progress updates.
type ActionTaskProgress struct {
	TaskID   string
	Task     Task
	Progress float64 // 0.0 to 1.0
	Message  string
}

// ActionTaskCompleted is a message sent when a task has completed.
type ActionTaskCompleted struct {
	TaskID string
	Task   Task
	Result any
	Error  error
}
