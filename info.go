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
	"os"
	"runtime"
	"time"
)

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
	i.values["OS"] = runtime.GOOS
	i.values["ARCH"] = runtime.GOARCH
	i.values["GO"] = runtime.Version()
	i.values["DS"] = string(os.PathSeparator)
	i.values["PS"] = string(os.PathListSeparator)
	i.values["PID"] = Itoa(os.Getpid())
	i.values["NUM_CPU"] = Itoa(runtime.NumCPU())
	i.values["USER"] = Username()

	if h, err := os.Hostname(); err == nil {
		i.values["HOSTNAME"] = h
	}

	// Directories — DS and DIR_HOME set first so Path() can use them.
	// CORE_HOME overrides os.UserHomeDir() (e.g., agent workspaces).
	if d := os.Getenv("CORE_HOME"); d != "" {
		i.values["DIR_HOME"] = d
	} else if d, err := os.UserHomeDir(); err == nil {
		i.values["DIR_HOME"] = d
	}

	// Derived directories via Path() — single point of responsibility
	i.values["DIR_DOWNLOADS"] = Path("Downloads")
	i.values["DIR_CODE"] = Path("Code")
	if d, err := os.UserConfigDir(); err == nil {
		i.values["DIR_CONFIG"] = d
	}
	if d, err := os.UserCacheDir(); err == nil {
		i.values["DIR_CACHE"] = d
	}
	i.values["DIR_TMP"] = os.TempDir()
	if d, err := os.Getwd(); err == nil {
		i.values["DIR_CWD"] = d
	}

	// Platform-specific data directory
	switch runtime.GOOS {
	case "darwin":
		i.values["DIR_DATA"] = Path(Env("DIR_HOME"), "Library")
	case "windows":
		if d := os.Getenv("LOCALAPPDATA"); d != "" {
			i.values["DIR_DATA"] = d
		}
	default:
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			i.values["DIR_DATA"] = xdg
		} else if Env("DIR_HOME") != "" {
			i.values["DIR_DATA"] = Path(Env("DIR_HOME"), ".local", "share")
		}
	}

	// Timestamps
	i.values["CORE_START"] = time.Now().UTC().Format(time.RFC3339)
}

// Env returns a system information value by key.
// Core keys (OS, DIR_HOME, DS, etc.) are pre-populated at init.
// Unknown keys fall through to os.Getenv — making Env a universal
// replacement for os.Getenv.
//
//	core.Env("OS")           // "darwin" (pre-populated)
//	core.Env("DIR_HOME")     // "/Users/snider" (pre-populated)
//	core.Env("FORGE_TOKEN")  // falls through to os.Getenv
func Env(key string) string {
	if v := systemInfo.values[key]; v != "" {
		return v
	}
	return os.Getenv(key)
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
