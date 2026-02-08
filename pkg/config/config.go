// Package config provides layered configuration management for the Core framework.
//
// Configuration values are resolved in priority order: defaults -> file -> env -> flags.
// Values are stored in a YAML file at ~/.core/config.yaml by default.
//
// Keys use dot notation for nested access:
//
//	cfg.Set("dev.editor", "vim")
//	var editor string
//	cfg.Get("dev.editor", &editor)
package config

import (
	"fmt"
	"os"
<<<<<<< HEAD
	"path/filepath"
=======
>>>>>>> fix/consolidate-workflows
	"strings"
	"sync"

	core "github.com/host-uk/core/pkg/framework/core"
<<<<<<< HEAD
	coreio "github.com/host-uk/core/pkg/io"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config implements the core.Config interface with layered resolution.
// It uses viper as the underlying configuration engine.
type Config struct {
	mu     sync.RWMutex
	v      *viper.Viper
	medium coreio.Medium
	path   string
=======
	"github.com/host-uk/core/pkg/io"
)

// Config implements the core.Config interface with layered resolution.
// Values are resolved in order: defaults -> file -> env -> flags.
type Config struct {
	mu     sync.RWMutex
	medium io.Medium
	path   string
	data   map[string]any
>>>>>>> fix/consolidate-workflows
}

// Option is a functional option for configuring a Config instance.
type Option func(*Config)

// WithMedium sets the storage medium for configuration file operations.
<<<<<<< HEAD
func WithMedium(m coreio.Medium) Option {
=======
func WithMedium(m io.Medium) Option {
>>>>>>> fix/consolidate-workflows
	return func(c *Config) {
		c.medium = m
	}
}

// WithPath sets the path to the configuration file.
func WithPath(path string) Option {
	return func(c *Config) {
		c.path = path
	}
}

<<<<<<< HEAD
// WithEnvPrefix sets the prefix for environment variables.
func WithEnvPrefix(prefix string) Option {
	return func(c *Config) {
		c.v.SetEnvPrefix(prefix)
	}
}

=======
>>>>>>> fix/consolidate-workflows
// New creates a new Config instance with the given options.
// If no medium is provided, it defaults to io.Local.
// If no path is provided, it defaults to ~/.core/config.yaml.
func New(opts ...Option) (*Config, error) {
	c := &Config{
<<<<<<< HEAD
		v: viper.New(),
	}

	// Configure viper defaults
	c.v.SetEnvPrefix("CORE_CONFIG")
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

=======
		data: make(map[string]any),
	}

>>>>>>> fix/consolidate-workflows
	for _, opt := range opts {
		opt(c)
	}

	if c.medium == nil {
<<<<<<< HEAD
		c.medium = coreio.Local
=======
		c.medium = io.Local
>>>>>>> fix/consolidate-workflows
	}

	if c.path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, core.E("config.New", "failed to determine home directory", err)
		}
<<<<<<< HEAD
		c.path = filepath.Join(home, ".core", "config.yaml")
	}

	c.v.AutomaticEnv()

	// Load existing config file if it exists
	if c.medium.Exists(c.path) {
		if err := c.LoadFile(c.medium, c.path); err != nil {
			return nil, core.E("config.New", "failed to load config file", err)
		}
=======
		c.path = home + "/.core/config.yaml"
	}

	// Load existing config file if it exists
	if c.medium.IsFile(c.path) {
		loaded, err := Load(c.medium, c.path)
		if err != nil {
			return nil, core.E("config.New", "failed to load config file", err)
		}
		c.data = loaded
	}

	// Overlay environment variables
	envData := LoadEnv("CORE_CONFIG_")
	for k, v := range envData {
		setNested(c.data, k, v)
>>>>>>> fix/consolidate-workflows
	}

	return c, nil
}

<<<<<<< HEAD
// LoadFile reads a configuration file from the given medium and path and merges it into the current config.
// It supports YAML and environment files (.env).
func (c *Config) LoadFile(m coreio.Medium, path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	content, err := m.Read(path)
	if err != nil {
		return core.E("config.LoadFile", "failed to read config file: "+path, err)
	}

	ext := filepath.Ext(path)
	if ext == "" && filepath.Base(path) == ".env" {
		c.v.SetConfigType("env")
	} else if ext != "" {
		c.v.SetConfigType(strings.TrimPrefix(ext, "."))
	} else {
		c.v.SetConfigType("yaml")
	}

	if err := c.v.MergeConfig(strings.NewReader(content)); err != nil {
		return core.E("config.LoadFile", "failed to parse config file: "+path, err)
	}

	return nil
}

// Get retrieves a configuration value by dot-notation key and stores it in out.
// If key is empty, it unmarshals the entire configuration into out.
// The out parameter must be a pointer to the target type.
=======
// Get retrieves a configuration value by dot-notation key and stores it in out.
// The out parameter must be a pointer to the target type.
// Returns an error if the key is not found.
>>>>>>> fix/consolidate-workflows
func (c *Config) Get(key string, out any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

<<<<<<< HEAD
	if key == "" {
		return c.v.Unmarshal(out)
	}

	if !c.v.IsSet(key) {
		return core.E("config.Get", fmt.Sprintf("key not found: %s", key), nil)
	}

	return c.v.UnmarshalKey(key, out)
=======
	val, ok := getNested(c.data, key)
	if !ok {
		return core.E("config.Get", fmt.Sprintf("key not found: %s", key), nil)
	}

	return assign(val, out)
>>>>>>> fix/consolidate-workflows
}

// Set stores a configuration value by dot-notation key and persists to disk.
func (c *Config) Set(key string, v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

<<<<<<< HEAD
	c.v.Set(key, v)

	// Persist to disk
	if err := Save(c.medium, c.path, c.v.AllSettings()); err != nil {
=======
	setNested(c.data, key, v)

	if err := Save(c.medium, c.path, c.data); err != nil {
>>>>>>> fix/consolidate-workflows
		return core.E("config.Set", "failed to save config", err)
	}

	return nil
}

// All returns a deep copy of all configuration values.
func (c *Config) All() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

<<<<<<< HEAD
	return c.v.AllSettings()
=======
	return deepCopyMap(c.data)
}

// deepCopyMap recursively copies a map[string]any.
func deepCopyMap(src map[string]any) map[string]any {
	result := make(map[string]any, len(src))
	for k, v := range src {
		switch val := v.(type) {
		case map[string]any:
			result[k] = deepCopyMap(val)
		case []any:
			cp := make([]any, len(val))
			copy(cp, val)
			result[k] = cp
		default:
			result[k] = v
		}
	}
	return result
>>>>>>> fix/consolidate-workflows
}

// Path returns the path to the configuration file.
func (c *Config) Path() string {
	return c.path
}

<<<<<<< HEAD
// Load reads a YAML configuration file from the given medium and path.
// Returns the parsed data as a map, or an error if the file cannot be read or parsed.
// Deprecated: Use Config.LoadFile instead.
func Load(m coreio.Medium, path string) (map[string]any, error) {
	content, err := m.Read(path)
	if err != nil {
		return nil, core.E("config.Load", "failed to read config file: "+path, err)
	}

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(content)); err != nil {
		return nil, core.E("config.Load", "failed to parse config file: "+path, err)
	}

	return v.AllSettings(), nil
}

// Save writes configuration data to a YAML file at the given path.
// It ensures the parent directory exists before writing.
func Save(m coreio.Medium, path string, data map[string]any) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return core.E("config.Save", "failed to marshal config", err)
	}

	dir := filepath.Dir(path)
	if err := m.EnsureDir(dir); err != nil {
		return core.E("config.Save", "failed to create config directory: "+dir, err)
	}

	if err := m.Write(path, string(out)); err != nil {
		return core.E("config.Save", "failed to write config file: "+path, err)
	}

=======
// getNested retrieves a value from a nested map using dot-notation keys.
func getNested(data map[string]any, key string) (any, bool) {
	parts := strings.Split(key, ".")
	current := any(data)

	for i, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		val, exists := m[part]
		if !exists {
			return nil, false
		}
		if i == len(parts)-1 {
			return val, true
		}
		current = val
	}

	return nil, false
}

// setNested sets a value in a nested map using dot-notation keys,
// creating intermediate maps as needed.
func setNested(data map[string]any, key string, value any) {
	parts := strings.Split(key, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}
		next, ok := current[part]
		if !ok {
			next = make(map[string]any)
			current[part] = next
		}
		m, ok := next.(map[string]any)
		if !ok {
			m = make(map[string]any)
			current[part] = m
		}
		current = m
	}
}

// assign sets the value of out to val, handling type conversions.
func assign(val any, out any) error {
	switch ptr := out.(type) {
	case *string:
		switch v := val.(type) {
		case string:
			*ptr = v
		default:
			*ptr = fmt.Sprintf("%v", v)
		}
	case *int:
		switch v := val.(type) {
		case int:
			*ptr = v
		case float64:
			*ptr = int(v)
		case int64:
			*ptr = int(v)
		default:
			return core.E("config.assign", fmt.Sprintf("cannot assign %T to *int", val), nil)
		}
	case *bool:
		switch v := val.(type) {
		case bool:
			*ptr = v
		default:
			return core.E("config.assign", fmt.Sprintf("cannot assign %T to *bool", val), nil)
		}
	case *float64:
		switch v := val.(type) {
		case float64:
			*ptr = v
		case int:
			*ptr = float64(v)
		case int64:
			*ptr = float64(v)
		default:
			return core.E("config.assign", fmt.Sprintf("cannot assign %T to *float64", val), nil)
		}
	case *any:
		*ptr = val
	case *map[string]any:
		switch v := val.(type) {
		case map[string]any:
			*ptr = v
		default:
			return core.E("config.assign", fmt.Sprintf("cannot assign %T to *map[string]any", val), nil)
		}
	default:
		return core.E("config.assign", fmt.Sprintf("unsupported target type: %T", out), nil)
	}
>>>>>>> fix/consolidate-workflows
	return nil
}

// Ensure Config implements core.Config at compile time.
var _ core.Config = (*Config)(nil)
