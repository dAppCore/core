// SPDX-License-Identifier: EUPL-1.2

// Settings, feature flags, and typed configuration for the Core framework.

package core

import (
	"sync"
)

// ConfigVar is a variable that can be set, unset, and queried for its state.
type ConfigVar[T any] struct {
	val T
	set bool
}

func (v *ConfigVar[T]) Get() T      { return v.val }
func (v *ConfigVar[T]) Set(val T)   { v.val = val; v.set = true }
func (v *ConfigVar[T]) IsSet() bool { return v.set }
func (v *ConfigVar[T]) Unset() {
	v.set = false
	var zero T
	v.val = zero
}

func NewConfigVar[T any](val T) ConfigVar[T] {
	return ConfigVar[T]{val: val, set: true}
}

// ConfigOptions holds configuration data.
type ConfigOptions struct {
	Settings map[string]any
	Features map[string]bool
}

func (o *ConfigOptions) init() {
	if o.Settings == nil {
		o.Settings = make(map[string]any)
	}
	if o.Features == nil {
		o.Features = make(map[string]bool)
	}
}

// Config holds configuration settings and feature flags.
type Config struct {
	*ConfigOptions
	mu sync.RWMutex
}

// Set stores a configuration value by key.
func (e *Config) Set(key string, val any) {
	e.mu.Lock()
	if e.ConfigOptions == nil {
		e.ConfigOptions = &ConfigOptions{}
	}
	e.ConfigOptions.init()
	e.Settings[key] = val
	e.mu.Unlock()
}

// Get retrieves a configuration value by key.
func (e *Config) Get(key string) Result {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.ConfigOptions == nil || e.Settings == nil {
		return Result{}
	}
	val, ok := e.Settings[key]
	if !ok {
		return Result{}
	}
	return Result{val, true}
}

func (e *Config) String(key string) string { return ConfigGet[string](e, key) }
func (e *Config) Int(key string) int       { return ConfigGet[int](e, key) }
func (e *Config) Bool(key string) bool     { return ConfigGet[bool](e, key) }

// ConfigGet retrieves a typed configuration value.
func ConfigGet[T any](e *Config, key string) T {
	r := e.Get(key)
	if !r.OK {
		var zero T
		return zero
	}
	typed, _ := r.Value.(T)
	return typed
}

// --- Feature Flags ---

func (e *Config) Enable(feature string) {
	e.mu.Lock()
	if e.ConfigOptions == nil {
		e.ConfigOptions = &ConfigOptions{}
	}
	e.ConfigOptions.init()
	e.Features[feature] = true
	e.mu.Unlock()
}

func (e *Config) Disable(feature string) {
	e.mu.Lock()
	if e.ConfigOptions == nil {
		e.ConfigOptions = &ConfigOptions{}
	}
	e.ConfigOptions.init()
	e.Features[feature] = false
	e.mu.Unlock()
}

func (e *Config) Enabled(feature string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.ConfigOptions == nil || e.Features == nil {
		return false
	}
	return e.Features[feature]
}

func (e *Config) EnabledFeatures() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.ConfigOptions == nil || e.Features == nil {
		return nil
	}
	var result []string
	for k, v := range e.Features {
		if v {
			result = append(result, k)
		}
	}
	return result
}
