// SPDX-License-Identifier: EUPL-1.2

// Core primitives: Option, Options, Result.
//
// Option is a single key-value pair. Options is a collection.
// Any function that returns Result can accept Options.
//
// Create options:
//
//	opts := core.Options{
//	    {K: "name", V: "brain"},
//	    {K: "path", V: "prompts"},
//	}
//
// Read options:
//
//	name := opts.String("name")
//	port := opts.Int("port")
//	ok := opts.Has("debug")
//
// Use with subsystems:
//
//	c.Drive().New(core.Options{
//	    {K: "name", V: "brain"},
//	    {K: "source", V: brainFS},
//	    {K: "path", V: "prompts"},
//	})
//
// Use with New:
//
//	c := core.New(core.Options{
//	    {K: "name", V: "myapp"},
//	})
package core

// Result is the universal return type for Core operations.
// Replaces the (value, error) pattern — errors flow through Core internally.
//
//	r := c.Data().New(core.Options{{K: "name", V: "brain"}})
//	if r.OK { use(r.Value) }
type Result[T any] struct {
	Value T
	OK    bool
}

// Option is a single key-value configuration pair.
//
//	core.Option{K: "name", V: "brain"}
//	core.Option{K: "port", V: 8080}
type Option struct {
	K string
	V any
}

// Options is a collection of Option items.
// The universal input type for Core operations.
//
//	opts := core.Options{{K: "name", V: "myapp"}}
//	name := opts.String("name")
type Options []Option

// Get retrieves a value by key.
//
//	val, ok := opts.Get("name")
func (o Options) Get(key string) (any, bool) {
	for _, opt := range o {
		if opt.K == key {
			return opt.V, true
		}
	}
	return nil, false
}

// Has returns true if a key exists.
//
//	if opts.Has("debug") { ... }
func (o Options) Has(key string) bool {
	_, ok := o.Get(key)
	return ok
}

// String retrieves a string value, empty string if missing.
//
//	name := opts.String("name")
func (o Options) String(key string) string {
	val, ok := o.Get(key)
	if !ok {
		return ""
	}
	s, _ := val.(string)
	return s
}

// Int retrieves an int value, 0 if missing.
//
//	port := opts.Int("port")
func (o Options) Int(key string) int {
	val, ok := o.Get(key)
	if !ok {
		return 0
	}
	i, _ := val.(int)
	return i
}

// Bool retrieves a bool value, false if missing.
//
//	debug := opts.Bool("debug")
func (o Options) Bool(key string) bool {
	val, ok := o.Get(key)
	if !ok {
		return false
	}
	b, _ := val.(bool)
	return b
}
