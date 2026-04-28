// SPDX-License-Identifier: EUPL-1.2

// Settings, feature flags, and typed configuration for the Core framework.

package core

import (
)

// ConfigVar is a variable that can be set, unset, and queried for its state.
//
//	host := core.NewConfigVar("homelab.lthn.sh")
//	if host.IsSet() { core.Println(host.Get()) }
type ConfigVar[T any] struct {
	val T
	set bool
}

// Get returns the current value.
//
//	val := v.Get()
func (v *ConfigVar[T]) Get() T { return v.val }

// Set sets the value and marks it as explicitly set.
//
//	v.Set(true)
func (v *ConfigVar[T]) Set(val T) { v.val = val; v.set = true }

// IsSet returns true if the value was explicitly set (distinguishes "set to false" from "never set").
//
//	if v.IsSet() { /* explicitly configured */ }
func (v *ConfigVar[T]) IsSet() bool { return v.set }

// Unset resets to zero value and marks as not set.
//
//	v.Unset()
//	v.IsSet()  // false
func (v *ConfigVar[T]) Unset() {
	v.set = false
	var zero T
	v.val = zero
}

// NewConfigVar creates a ConfigVar with an initial value marked as set.
//
//	debug := core.NewConfigVar(true)
func NewConfigVar[T any](val T) ConfigVar[T] {
	return ConfigVar[T]{val: val, set: true}
}

// ConfigOptions holds configuration data.
//
//	opts := core.ConfigOptions{
//	    Settings: map[string]any{"config.host": "homelab.lthn.sh"},
//	    Features: map[string]bool{"agentic": true},
//	}
//	_ = opts
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
//
//	cfg := (&core.Config{}).New()
//	cfg.Set("config.host", "homelab.lthn.sh")
//	core.Println(cfg.String("config.host"))
type Config struct {
	*ConfigOptions
	mu RWMutex
}

// New initialises a Config with empty settings and features.
//
//	cfg := (&core.Config{}).New()
func (e *Config) New() *Config {
	e.ConfigOptions = &ConfigOptions{}
	e.ConfigOptions.init()
	return e
}

// Set stores a configuration value by key.
//
//	cfg := (&core.Config{}).New()
//	cfg.Set("config.host", "homelab.lthn.sh")
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
//
//	cfg := (&core.Config{}).New()
//	cfg.Set("config.host", "homelab.lthn.sh")
//	r := cfg.Get("config.host")
//	if r.OK { core.Println(r.Value.(string)) }
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

// String retrieves a string config value (empty string if missing).
//
//	host := c.Config().String("database.host")
func (e *Config) String(key string) string { return ConfigGet[string](e, key) }

// Int retrieves an int config value (0 if missing).
//
//	port := c.Config().Int("database.port")
func (e *Config) Int(key string) int { return ConfigGet[int](e, key) }

// Bool retrieves a bool config value (false if missing).
//
//	debug := c.Config().Bool("debug")
func (e *Config) Bool(key string) bool { return ConfigGet[bool](e, key) }

// ConfigGet retrieves a typed configuration value.
//
//	cfg := (&core.Config{}).New()
//	cfg.Set("config.host", "homelab.lthn.sh")
//	host := core.ConfigGet[string](cfg, "config.host")
//	core.Println(host)
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

// Enable activates a feature flag.
//
//	c.Config().Enable("dark-mode")
func (e *Config) Enable(feature string) {
	e.mu.Lock()
	if e.ConfigOptions == nil {
		e.ConfigOptions = &ConfigOptions{}
	}
	e.ConfigOptions.init()
	e.Features[feature] = true
	e.mu.Unlock()
}

// Disable deactivates a feature flag.
//
//	c.Config().Disable("dark-mode")
func (e *Config) Disable(feature string) {
	e.mu.Lock()
	if e.ConfigOptions == nil {
		e.ConfigOptions = &ConfigOptions{}
	}
	e.ConfigOptions.init()
	e.Features[feature] = false
	e.mu.Unlock()
}

// Enabled returns true if a feature flag is active.
//
//	if c.Config().Enabled("dark-mode") { ... }
func (e *Config) Enabled(feature string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.ConfigOptions == nil || e.Features == nil {
		return false
	}
	return e.Features[feature]
}

// EnabledFeatures returns all active feature flag names.
//
//	features := c.Config().EnabledFeatures()
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
