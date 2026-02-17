package coredeno

import (
	"path/filepath"
	"strings"
)

// CheckPath returns true if the given path is under any of the allowed prefixes.
// Empty allowed list means deny all (secure by default).
func CheckPath(path string, allowed []string) bool {
	if len(allowed) == 0 {
		return false
	}
	clean := filepath.Clean(path)
	for _, prefix := range allowed {
		cleanPrefix := filepath.Clean(prefix)
		if strings.HasPrefix(clean, cleanPrefix) {
			return true
		}
	}
	return false
}

// CheckNet returns true if the given host:port is in the allowed list.
func CheckNet(addr string, allowed []string) bool {
	for _, a := range allowed {
		if a == addr {
			return true
		}
	}
	return false
}

// CheckRun returns true if the given command is in the allowed list.
func CheckRun(cmd string, allowed []string) bool {
	for _, a := range allowed {
		if a == cmd {
			return true
		}
	}
	return false
}
