// SPDX-License-Identifier: EUPL-1.2

// Core primitives: Option, Options, Result.
//
// Option is a single key-value pair. Options is a collection.
// Any function that returns Result can accept Options.
//
// Create options:
//
//	opts := core.Options{
//	    {Key: "name", Value: "brain"},
//	    {Key: "path", Value: "prompts"},
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
//	    {Key: "name", Value: "brain"},
//	    {Key: "source", Value: brainFS},
//	    {Key: "path", Value: "prompts"},
//	})
//
// Use with New:
//
//	c := core.New(core.Options{
//	    {Key: "name", Value: "myapp"},
//	})
package core

// Result is the universal return type for Core operations.
// Replaces the (value, error) pattern — errors flow through Core internally.
//
//	r := c.Data().New(core.Options{{Key: "name", Value: "brain"}})
//	if r.OK { use(r.Result()) }
type Result struct {
	Value any
	OK    bool
}

// Result gets or sets the value. Zero args returns Value. With args, maps
// Go (value, error) pairs to Result and returns self.
//
//	r.Result(file, err)     // OK = err == nil, Value = file
//	r.Result(value)         // OK = true, Value = value
//	r.Result()              // after set — returns the value
func (r Result) Result(args ...any) Result {

	if len(args) == 1 {
		return Result{args[0], true}
	}

	if len(args) >= 2 {
		if err, ok := args[len(args)-1].(error); ok {
			if err != nil {
				return Result{err, false}
			}
			return Result{args[0], true}
		}
	}
	return Result{args[0], true}

}

// Option is a single key-value configuration pair.
//
//	core.Option{Key: "name", Value: "brain"}
//	core.Option{Key: "port", Value: 8080}
type Option struct {
	Key   string
	Value any
}

// Options is a collection of Option items.
// The universal input type for Core operations.
//
//	opts := core.Options{{Key: "name", Value: "myapp"}}
//	name := opts.String("name")
type Options []Option

// Get retrieves a value by key.
//
//	r := opts.Get("name")
//	if r.OK { name := r.Value.(string) }
func (o Options) Get(key string) Result {
	for _, opt := range o {
		if opt.Key == key {
			return Result{opt.Value, true}
		}
	}
	return Result{}
}

// Has returns true if a key exists.
//
//	if opts.Has("debug") { ... }
func (o Options) Has(key string) bool {
	return o.Get(key).OK
}

// String retrieves a string value, empty string if missing.
//
//	name := opts.String("name")
func (o Options) String(key string) string {
	r := o.Get(key)
	if !r.OK {
		return ""
	}
	s, _ := r.Value.(string)
	return s
}

// Int retrieves an int value, 0 if missing.
//
//	port := opts.Int("port")
func (o Options) Int(key string) int {
	r := o.Get(key)
	if !r.OK {
		return 0
	}
	i, _ := r.Value.(int)
	return i
}

// Bool retrieves a bool value, false if missing.
//
//	debug := opts.Bool("debug")
func (o Options) Bool(key string) bool {
	r := o.Get(key)
	if !r.OK {
		return false
	}
	b, _ := r.Value.(bool)
	return b
}
