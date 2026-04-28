// SPDX-License-Identifier: EUPL-1.2

// Base OS primitives for the Core framework.
//
// Re-exports stdlib os value types and the standard streams so consumer
// packages declare FileMode parameters and reach Stdin/Stdout/Stderr
// without importing os directly.
//
// File operations live on c.Fs() (sandbox-aware). Environment lookups
// live on core.Env. Process control lives on c.Process() / go-process.
// What's here is the connecting tissue: types, constants, and the
// canonical stdio streams that boundary code can't avoid.
//
// Usage:
//
//	func writeMode(p string, mode core.FileMode) { ... }
//
//	core.WriteString(core.Stderr(), "diagnostic\n")
//
//	if mode.Perm()&core.ModePerm == 0o600 { ... }
package core

import "os"

// OSFile is an alias for os.File.
//
//	r := core.Create("agent.log")
//	if r.OK { file := r.Value.(*core.OSFile); _ = file }
type OSFile = os.File

// FileMode is an alias for os.FileMode — file mode bits and permissions.
//
//	mode := core.FileMode(0o600)
//	core.Println(mode.Perm())
type FileMode = os.FileMode

// File mode bits exposed at core scope. These are the same values as
// os.ModeDir etc., re-exported so consumers don't need to import os.
//
//	mode := core.ModeDir | core.ModePerm
//	if mode&core.ModeDir != 0 { core.Println("directory") }
const (
	ModeDir        = os.ModeDir
	ModeAppend     = os.ModeAppend
	ModeExclusive  = os.ModeExclusive
	ModeTemporary  = os.ModeTemporary
	ModeSymlink    = os.ModeSymlink
	ModeDevice     = os.ModeDevice
	ModeNamedPipe  = os.ModeNamedPipe
	ModeSocket     = os.ModeSocket
	ModeSetuid     = os.ModeSetuid
	ModeSetgid     = os.ModeSetgid
	ModeCharDevice = os.ModeCharDevice
	ModeSticky     = os.ModeSticky
	ModeIrregular  = os.ModeIrregular
	ModeType       = os.ModeType
	ModePerm       = os.ModePerm // 0o777 — Unix permission bits
)

// File open flags exposed at core scope.
//
//	r := core.OpenFile("agent.log", core.O_CREATE|core.O_WRONLY, 0o644)
//	if !r.OK { return r }
const (
	O_APPEND = os.O_APPEND
	O_CREATE = os.O_CREATE
	O_EXCL   = os.O_EXCL
	O_RDONLY = os.O_RDONLY
	O_RDWR   = os.O_RDWR
	O_SYNC   = os.O_SYNC
	O_TRUNC  = os.O_TRUNC
	O_WRONLY = os.O_WRONLY
)

// Path separators exposed at core scope.
//
//	parts := core.Split("a/b", string(core.PathSeparator))
const (
	PathSeparator     = os.PathSeparator
	PathListSeparator = os.PathListSeparator
)

// (Note: core.Signal is the existing Core primitive in signal.go for
// signal-event handling — distinct from os.Signal the interface. Use
// c.Signal() for the action-based signal surface.)

// Stdin returns the canonical standard input stream as an io.Reader.
//
//	scanner := core.NewLineScanner(core.Stdin())
func Stdin() Reader {
	return os.Stdin
}

// Stdout returns the canonical standard output stream as an io.Writer.
//
//	core.WriteString(core.Stdout(), "ready\n")
func Stdout() Writer {
	return os.Stdout
}

// Stderr returns the canonical standard error stream as an io.Writer.
//
//	core.WriteString(core.Stderr(), "warning\n")
func Stderr() Writer {
	return os.Stderr
}

// ReadFile reads the named file and returns its bytes.
//
//	r := core.ReadFile("config/agent.json")
//	if r.OK { data := r.Value.([]byte); _ = data }
func ReadFile(p string) Result {
	data, err := os.ReadFile(p)
	if err != nil {
		return Result{err, false}
	}
	return Result{data, true}
}

// WriteFile writes data to the named file with mode.
//
//	r := core.WriteFile("config/agent.json", []byte("{}"), 0o644)
func WriteFile(p string, data []byte, mode FileMode) Result {
	if err := os.WriteFile(p, data, mode); err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}

// MkdirAll creates a directory path and all missing parents.
//
//	r := core.MkdirAll("logs/agent", 0o755)
func MkdirAll(p string, mode FileMode) Result {
	if err := os.MkdirAll(p, mode); err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}

// Mkdir creates one directory.
//
//	r := core.Mkdir("logs", 0o755)
func Mkdir(p string, mode FileMode) Result {
	if err := os.Mkdir(p, mode); err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}

// Create creates or truncates the named file.
//
//	r := core.Create("logs/agent.log")
//	if r.OK { file := r.Value.(*core.OSFile); _ = file }
func Create(p string) Result {
	return Result{}.New(os.Create(p))
}

// Open opens the named file for reading.
//
//	r := core.Open("config/agent.json")
func Open(p string) Result {
	return Result{}.New(os.Open(p))
}

// OpenFile opens the named file with explicit flags and mode.
//
//	r := core.OpenFile("logs/agent.log", core.O_APPEND|core.O_CREATE|core.O_WRONLY, 0o644)
func OpenFile(p string, flag int, mode FileMode) Result {
	return Result{}.New(os.OpenFile(p, flag, mode))
}

// Stat returns file information for p.
//
//	r := core.Stat("config/agent.json")
func Stat(p string) Result {
	return Result{}.New(os.Stat(p))
}

// Lstat returns file information for p without following a final symlink.
//
//	r := core.Lstat("config/current")
func Lstat(p string) Result {
	return Result{}.New(os.Lstat(p))
}

// Remove removes the named file or empty directory.
//
//	r := core.Remove("logs/old-agent.log")
func Remove(p string) Result {
	if err := os.Remove(p); err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}

// RemoveAll removes a path and any children.
//
//	r := core.RemoveAll("tmp/session-42")
func RemoveAll(p string) Result {
	if err := os.RemoveAll(p); err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}

// Rename renames oldPath to newPath.
//
//	r := core.Rename("config.tmp", "config.json")
func Rename(oldPath, newPath string) Result {
	if err := os.Rename(oldPath, newPath); err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}

// MkdirTemp creates a new temporary directory.
//
//	r := core.MkdirTemp("", "agent-*")
func MkdirTemp(dir, pattern string) Result {
	return Result{}.New(os.MkdirTemp(dir, pattern))
}

// TempDir returns the default directory for temporary files.
//
//	dir := core.TempDir()
func TempDir() string {
	return os.TempDir()
}

// IsNotExist reports whether err indicates a missing path.
//
//	if core.IsNotExist(err) { core.Println("missing") }
func IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

// IsExist reports whether err indicates an existing path.
//
//	if core.IsExist(err) { core.Println("exists") }
func IsExist(err error) bool {
	return os.IsExist(err)
}

// IsPermission reports whether err indicates a permission failure.
//
//	if core.IsPermission(err) { core.Println("denied") }
func IsPermission(err error) bool {
	return os.IsPermission(err)
}

// DirFS returns an FS rooted at the given directory path.
//
//	fsys := core.DirFS("/path/to/templates")
func DirFS(dir string) FS {
	return os.DirFS(dir)
}

// Args returns the command-line arguments.
//
//	args := core.Args()
func Args() []string {
	return os.Args
}

// Hostname returns the kernel host name.
//
//	r := core.Hostname()
func Hostname() Result {
	return Result{}.New(os.Hostname())
}

// Getpid returns the process id of the caller.
//
//	pid := core.Getpid()
func Getpid() int {
	return os.Getpid()
}

// Getppid returns the parent process id of the caller.
//
//	ppid := core.Getppid()
func Getppid() int {
	return os.Getppid()
}

// Getwd returns the current working directory.
//
//	r := core.Getwd()
func Getwd() Result {
	return Result{}.New(os.Getwd())
}

// Chdir changes the current working directory.
//
//	r := core.Chdir("/tmp")
func Chdir(dir string) Result {
	if err := os.Chdir(dir); err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}

// UserHomeDir returns the current user's home directory.
//
//	r := core.UserHomeDir()
func UserHomeDir() Result {
	return Result{}.New(os.UserHomeDir())
}

// UserConfigDir returns the default root directory for user configuration.
//
//	r := core.UserConfigDir()
func UserConfigDir() Result {
	return Result{}.New(os.UserConfigDir())
}

// UserCacheDir returns the default root directory for user cache data.
//
//	r := core.UserCacheDir()
func UserCacheDir() Result {
	return Result{}.New(os.UserCacheDir())
}

// Environ returns a copy of strings representing the environment.
//
//	env := core.Environ()
func Environ() []string {
	return os.Environ()
}

// Getenv retrieves the value of the environment variable named by key.
//
//	token := core.Getenv("FORGE_TOKEN")
func Getenv(key string) string {
	return os.Getenv(key)
}

// Setenv sets an environment variable.
//
//	err := core.Setenv("FORGE_TOKEN", token)
func Setenv(key, value string) error {
	return os.Setenv(key, value)
}

// Unsetenv removes an environment variable.
//
//	err := core.Unsetenv("FORGE_TOKEN")
func Unsetenv(key string) error {
	return os.Unsetenv(key)
}

// LookupEnv retrieves the value of the environment variable named by key.
//
//	value, ok := core.LookupEnv("FORGE_TOKEN")
func LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

// osExit is the test hook for process termination. Production wires it to
// os.Exit here so os.go remains the sole production os owner.
var osExit = os.Exit

// Exit terminates the current process with code.
//
//	core.Exit(1)
func Exit(code int) {
	osExit(code)
}
