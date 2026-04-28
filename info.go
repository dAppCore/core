// SPDX-License-Identifier: EUPL-1.2

// System information registry for the Core framework.
// Read-only key-value store of environment facts, populated once at init.
// Env is environment. Config is ours.
//
// System keys:
//
//	core.Env("OS")        // "darwin"
//	core.Env("ARCH")      // "arm64"
//	core.Env("GO")        // "go1.26"
//	core.Env("DS")        // "/" (directory separator)
//	core.Env("PS")        // ":" (path list separator)
//	core.Env("HOSTNAME")  // "cladius"
//	core.Env("USER")      // "snider"
//	core.Env("PID")       // "12345"
//	core.Env("NUM_CPU")   // "10"
//
// Directory keys:
//
//	core.Env("DIR_HOME")      // "/Users/snider"
//	core.Env("DIR_CONFIG")    // "~/Library/Application Support"
//	core.Env("DIR_CACHE")     // "~/Library/Caches"
//	core.Env("DIR_DATA")      // "~/Library" (platform-specific)
//	core.Env("DIR_TMP")       // "/tmp"
//	core.Env("DIR_CWD")       // current working directory
//	core.Env("DIR_DOWNLOADS") // "~/Downloads"
//	core.Env("DIR_CODE")      // "~/Code"
//
// Timestamp keys:
//
//	core.Env("CORE_START")  // "2026-03-22T14:30:00Z"
package core

import (
	"runtime"
	"runtime/debug"
)

// OS returns the operating system name (darwin, linux, windows, etc.)
// the binary is running on.
//
//	if core.OS() == "darwin" { core.Println("on a Mac") }
func OS() string {
	return runtime.GOOS
}

// Arch returns the CPU architecture (amd64, arm64, etc.) the binary
// was compiled for.
//
//	if core.Arch() == "arm64" { core.Println("Apple Silicon or arm64 Linux") }
func Arch() string {
	return runtime.GOARCH
}

// GoVersion returns the Go runtime version string (e.g. "go1.26.0")
// the binary was compiled with.
//
//	core.Println(core.GoVersion())  // "go1.26.0"
func GoVersion() string {
	return runtime.Version()
}

// NumCPU returns the number of logical CPUs available to the running
// process.
//
//	workers := core.NumCPU()
//	core.Println(workers)
func NumCPU() int {
	return runtime.NumCPU()
}

// StackBuf returns the Go runtime stack trace of the current goroutine
// as a byte slice. Used for crash reports and panic recovery.
//
//	report := core.StackBuf()
//	core.Println(string(report))
func StackBuf() []byte {
	return debug.Stack()
}

// SysInfo holds read-only system information, populated once at init.
//
//	home := core.Env("DIR_HOME")
//	core.Println(home)
type SysInfo struct {
	values map[string]string
}

// systemInfo is declared empty — populated in init() so Path() can be used
// without creating an init cycle.
var systemInfo = &SysInfo{values: make(map[string]string)}

func init() {
	i := systemInfo

	// System
	i.values["OS"] = OS()
	i.values["ARCH"] = Arch()
	i.values["GO"] = GoVersion()
	i.values["DS"] = string(PathSeparator)
	i.values["PS"] = string(PathListSeparator)
	i.values["PID"] = Itoa(Getpid())
	i.values["NUM_CPU"] = Itoa(NumCPU())
	i.values["USER"] = Username()

	if r := Hostname(); r.OK {
		i.values["HOSTNAME"] = r.Value.(string)
	}

	// Directories — DS and DIR_HOME set first so Path() can use them.
	// CORE_HOME overrides core.UserHomeDir() (e.g., agent workspaces).
	if d := Getenv("CORE_HOME"); d != "" {
		i.values["DIR_HOME"] = d
	} else if r := UserHomeDir(); r.OK {
		i.values["DIR_HOME"] = r.Value.(string)
	}

	// Derived directories via Path() — single point of responsibility
	i.values["DIR_DOWNLOADS"] = Path("Downloads")
	i.values["DIR_CODE"] = Path("Code")
	if r := UserConfigDir(); r.OK {
		i.values["DIR_CONFIG"] = r.Value.(string)
	}
	if r := UserCacheDir(); r.OK {
		i.values["DIR_CACHE"] = r.Value.(string)
	}
	i.values["DIR_TMP"] = TempDir()
	if r := Getwd(); r.OK {
		i.values["DIR_CWD"] = r.Value.(string)
	}

	// Platform-specific data directory
	switch OS() {
	case "darwin":
		i.values["DIR_DATA"] = Path(Env("DIR_HOME"), "Library")
	case "windows":
		if d := Getenv("LOCALAPPDATA"); d != "" {
			i.values["DIR_DATA"] = d
		}
	default:
		if xdg := Getenv("XDG_DATA_HOME"); xdg != "" {
			i.values["DIR_DATA"] = xdg
		} else if Env("DIR_HOME") != "" {
			i.values["DIR_DATA"] = Path(Env("DIR_HOME"), ".local", "share")
		}
	}

	// Timestamps
	i.values["CORE_START"] = Now().UTC().Format(TimeRFC3339)
}

// Env returns a system information value by key.
// Core keys (OS, DIR_HOME, DS, etc.) are pre-populated at init.
// Unknown keys fall through to core.Getenv — making Env a universal
// replacement for core.Getenv.
//
//	core.Env("OS")           // "darwin" (pre-populated)
//	core.Env("DIR_HOME")     // "/Users/snider" (pre-populated)
//	core.Env("FORGE_TOKEN")  // falls through to core.Getenv
func Env(key string) string {
	if v := systemInfo.values[key]; v != "" {
		return v
	}
	return Getenv(key)
}

// EnvKeys returns all available environment keys.
//
//	keys := core.EnvKeys()
func EnvKeys() []string {
	keys := make([]string, 0, len(systemInfo.values))
	for k := range systemInfo.values {
		keys = append(keys, k)
	}
	return keys
}
