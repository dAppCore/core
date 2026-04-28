// SPDX-License-Identifier: EUPL-1.2

// Environment mutation helpers for the Core framework.
// Wraps os environment mutation APIs so downstream packages can stay on core.
package core

// UnsetEnv removes an environment variable.
//
//	err := core.UnsetEnv("FORGE_TOKEN")
func UnsetEnv(key string) error {
	return Unsetenv(key)
}
