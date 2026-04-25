// SPDX-License-Identifier: EUPL-1.2

// Environment mutation helpers for the Core framework.
// Wraps os environment mutation APIs so downstream packages can stay on core.
package core

import "os"

// Setenv sets an environment variable using os.Setenv.
//
//	err := core.Setenv("FORGE_TOKEN", token)
func Setenv(key, value string) error {
	return os.Setenv(key, value)
}

// UnsetEnv removes an environment variable using os.Unsetenv.
//
//	err := core.UnsetEnv("FORGE_TOKEN")
func UnsetEnv(key string) error {
	return os.Unsetenv(key)
}
