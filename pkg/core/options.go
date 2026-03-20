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
//	if r.OK { use(r.Result()) }
type Result struct {
	Value any
	OK    bool
}

// Result returns the value.
//
//	val := r.Result()
func (r *Result) Result() any { return r.Value }

// New creates a Result from variadic args.
// Maps Go (value, error) pairs to Result.
//
//	r.New(file, err)       // OK = err == nil, Value = file
//	r.New(value)           // OK = true, Value = value
//	r.New()                // OK = false
func (r *Result) New(args ...any) *Result {
	if len(args) == 0 {
		r.OK = false
		return r
	}

	// Check if last arg is an error
	if len(args) >= 2 {
		if err, ok := args[len(args)-1].(error); ok {
			if err != nil {
				r.Value = err
				r.OK = false
				return r
			}
			r.Value = args[0]
			r.OK = true
			return r
		}
	}

	r.Value = args[0]
	r.OK = true
	return r
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
