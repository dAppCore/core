// SPDX-License-Identifier: EUPL-1.2

// Settings, feature flags, and typed configuration for the Core framework.
// Named after /etc — the configuration directory.

package core

import (
	"sync"
)

// Var is a variable that can be set, unset, and queried for its state.
// Zero value is unset.
type ConfigVar[T any] struct {
	val T
	set bool
}

// Get returns the value, or the zero value if unset.
func (v *ConfigVar[T]) Get() T { return v.val }

// Set sets the value and marks it as set.
func (v *ConfigVar[T]) Set(val T) { v.val = val; v.set = true }

// IsSet returns true when a value has been set.
func (v *ConfigVar[T]) IsSet() bool { return v.set }

// Unset resets to zero value and marks as unset.
func (v *ConfigVar[T]) Unset() {
	v.set = false
	var zero T
	v.val = zero
}

// NewVar creates a Var with the given value (marked as set).
func NewConfigVar[T any](val T) ConfigVar[T] {
	return ConfigVar[T]{val: val, set: true}
}

// Config holds configuration settings and feature flags.
type Config struct {
	mu       sync.RWMutex
	settings map[string]any
	features map[string]bool
}

// NewConfig creates a new configuration store.
func NewConfig() *Config {
	return &Config{
		settings: make(map[string]any),
		features: make(map[string]bool),
	}
}

// Set stores a configuration value by key.
func (e *Config) Set(key string, val any) {
	e.mu.Lock()
	e.settings[key] = val
	e.mu.Unlock()
}

// Get retrieves a configuration value by key.
// Returns (value, true) if found, (zero, false) if not.
func (e *Config) Get(key string) (any, bool) {
	e.mu.RLock()
	val, ok := e.settings[key]
	e.mu.RUnlock()
	return val, ok
}

func (e *Config) String(key string) string { return ConfigGet[string](e, key) }
func (e *Config) Int(key string) int       { return ConfigGet[int](e, key) }
func (e *Config) Bool(key string) bool     { return ConfigGet[bool](e, key) }

// ConfigGet retrieves a typed configuration value.
// Returns zero value if key is missing or type doesn't match.
func ConfigGet[T any](e *Config, key string) T {
	val, ok := e.Get(key)
	if !ok {
		var zero T
		return zero
	}
	typed, _ := val.(T)
	return typed
}

// --- Feature Flags ---

// Enable enables a feature flag.
func (e *Config) Enable(feature string) {
	e.mu.Lock()
	e.features[feature] = true
	e.mu.Unlock()
}

// Disable disables a feature flag.
func (e *Config) Disable(feature string) {
	e.mu.Lock()
	e.features[feature] = false
	e.mu.Unlock()
}

// Enabled returns true if the feature is enabled.
func (e *Config) Enabled(feature string) bool {
	e.mu.RLock()
	v := e.features[feature]
	e.mu.RUnlock()
	return v
}

// Features returns all enabled feature names.
func (e *Config) EnabledFeatures() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	var result []string
	for k, v := range e.features {
		if v {
			result = append(result, k)
		}
	}
	return result
}
