// SPDX-License-Identifier: EUPL-1.2

// Core primitives: Option, Options, Result.
//
// Options is the universal input type. Result is the universal output type.
// All Core operations accept Options and return Result.
//
//	opts := core.NewOptions(
//	    core.Option{Key: "name", Value: "brain"},
//	    core.Option{Key: "path", Value: "prompts"},
//	)
//	r := c.Drive().New(opts)
//	if !r.OK { log.Fatal(r.Error()) }
package core

// --- Result: Universal Output ---

// Result is the universal return type for Core operations.
// Replaces the (value, error) pattern — errors flow through Core internally.
//
//	r := c.Data().New(opts)
//	if !r.OK { core.Error("failed", "err", r.Error()) }
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
	if len(args) == 0 {
		return r
	}
	return r.New(args...)
}

// New adapts Go (value, error) pairs into a Result.
//
//	r := core.Result{}.New(file, err)
func (r Result) New(args ...any) Result {
	if len(args) == 0 {
		return r
	}

	if len(args) > 1 {
		if err, ok := args[len(args)-1].(error); ok {
			if err != nil {
				return Result{Value: err, OK: false}
			}
			r.Value = args[0]
			r.OK = true
			return r
		}
	}

	r.Value = args[0]

	if err, ok := r.Value.(error); ok {
		if err != nil {
			return Result{Value: err, OK: false}
		}
		return Result{OK: true}
	}

	r.OK = true
	return r
}

// Get returns the Result if OK, empty Result otherwise.
//
//	r := core.Result{Value: "hello", OK: true}.Get()
func (r Result) Get() Result {
	if r.OK {
		return r
	}
	return Result{Value: r.Value, OK: false}
}

// Option is a single key-value configuration pair.
//
//	core.Option{Key: "name", Value: "brain"}
//	core.Option{Key: "port", Value: 8080}
type Option struct {
	Key   string
	Value any
}

// --- Options: Universal Input ---

// Options is the universal input type for Core operations.
// A structured collection of key-value pairs with typed accessors.
//
//	opts := core.NewOptions(
//	    core.Option{Key: "name", Value: "myapp"},
//	    core.Option{Key: "port", Value: 8080},
//	)
//	name := opts.String("name")
type Options struct {
	items []Option
}

// NewOptions creates an Options collection from key-value pairs.
//
//	opts := core.NewOptions(
//	    core.Option{Key: "name", Value: "brain"},
//	    core.Option{Key: "path", Value: "prompts"},
//	)
func NewOptions(items ...Option) Options {
	cp := make([]Option, len(items))
	copy(cp, items)
	return Options{items: cp}
}

// Set adds or updates a key-value pair.
//
//	opts.Set("port", 8080)
func (o *Options) Set(key string, value any) {
	for i, opt := range o.items {
		if opt.Key == key {
			o.items[i].Value = value
			return
		}
	}
	o.items = append(o.items, Option{Key: key, Value: value})
}

// Get retrieves a value by key.
//
//	r := opts.Get("name")
//	if r.OK { name := r.Value.(string) }
func (o Options) Get(key string) Result {
	for _, opt := range o.items {
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

// Len returns the number of options.
func (o Options) Len() int {
	return len(o.items)
}

// Items returns a copy of the underlying option slice.
func (o Options) Items() []Option {
	cp := make([]Option, len(o.items))
	copy(cp, o.items)
	return cp
}
