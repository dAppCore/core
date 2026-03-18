// SPDX-License-Identifier: EUPL-1.2

// Settings, feature flags, and typed configuration for the Core framework.
// Named after /etc — the configuration directory.

package core

import (
	"sync"
)

// Var is a variable that can be set, unset, and queried for its state.
// Zero value is unset.
type Var[T any] struct {
	val T
	set bool
}

// Get returns the value, or the zero value if unset.
func (v *Var[T]) Get() T { return v.val }

// Set sets the value and marks it as set.
func (v *Var[T]) Set(val T) { v.val = val; v.set = true }

// IsSet returns true when a value has been set.
func (v *Var[T]) IsSet() bool { return v.set }

// Unset resets to zero value and marks as unset.
func (v *Var[T]) Unset() {
	v.set = false
	var zero T
	v.val = zero
}

// NewVar creates a Var with the given value (marked as set).
func NewVar[T any](val T) Var[T] {
	return Var[T]{val: val, set: true}
}

// Etc holds configuration settings and feature flags.
type Etc struct {
	mu       sync.RWMutex
	settings map[string]any
	features map[string]bool
}

// NewEtc creates a new configuration store.
func NewEtc() *Etc {
	return &Etc{
		settings: make(map[string]any),
		features: make(map[string]bool),
	}
}

// Set stores a configuration value by key.
func (e *Etc) Set(key string, val any) {
	e.mu.Lock()
	e.settings[key] = val
	e.mu.Unlock()
}

// Get retrieves a configuration value by key.
// Returns (value, true) if found, (zero, false) if not.
func (e *Etc) Get(key string) (any, bool) {
	e.mu.RLock()
	val, ok := e.settings[key]
	e.mu.RUnlock()
	return val, ok
}

// GetString retrieves a string configuration value.
func (e *Etc) GetString(key string) string { return EtcGet[string](e, key) }

// GetInt retrieves an int configuration value.
func (e *Etc) GetInt(key string) int { return EtcGet[int](e, key) }

// GetBool retrieves a bool configuration value.
func (e *Etc) GetBool(key string) bool { return EtcGet[bool](e, key) }

// EtcGet retrieves a typed configuration value.
// Returns zero value if key is missing or type doesn't match.
func EtcGet[T any](e *Etc, key string) T {
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
func (e *Etc) Enable(feature string) {
	e.mu.Lock()
	e.features[feature] = true
	e.mu.Unlock()
}

// Disable disables a feature flag.
func (e *Etc) Disable(feature string) {
	e.mu.Lock()
	e.features[feature] = false
	e.mu.Unlock()
}

// Enabled returns true if the feature is enabled.
func (e *Etc) Enabled(feature string) bool {
	e.mu.RLock()
	v := e.features[feature]
	e.mu.RUnlock()
	return v
}

// Features returns all enabled feature names.
func (e *Etc) EnabledFeatures() []string {
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
