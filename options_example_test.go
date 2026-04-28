package core_test

import . "dappco.re/go"

// ExampleResult_New creates a successful Result containing a created status for agent options.
// Options carry loosely typed inputs while typed accessors keep call sites small.
func ExampleResult_New() {
	r := Result{}.New("created", nil)
	Println(r.OK)
	Println(r.Value)
	// Output:
	// true
	// created
}

// ExampleOption declares a key-value option through `Option` for agent options. Options
// carry loosely typed inputs while typed accessors keep call sites small.
func ExampleOption() {
	opt := Option{Key: "port", Value: 8080}
	Println(opt.Key)
	Println(opt.Value)
	// Output:
	// port
	// 8080
}

// ExampleNewOptions_withItems loads initial items through `NewOptions` for agent options.
// Options carry loosely typed inputs while typed accessors keep call sites small.
func ExampleNewOptions_withItems() {
	opts := NewOptions(
		Option{Key: "name", Value: "api"},
		Option{Key: "port", Value: 8080},
	)
	Println(opts.Len())
	// Output: 2
}

// ExampleOptions_Set sets a value through `Options.Set` for agent options. Options carry
// loosely typed inputs while typed accessors keep call sites small.
func ExampleOptions_Set() {
	opts := NewOptions()
	opts.Set("name", "api")
	opts.Set("name", "worker")
	Println(opts.String("name"))
	Println(opts.Len())
	// Output:
	// worker
	// 1
}

// ExampleOptions_Get retrieves a value through `Options.Get` for agent options. Options
// carry loosely typed inputs while typed accessors keep call sites small.
func ExampleOptions_Get() {
	opts := NewOptions(Option{Key: "name", Value: "api"})
	r := opts.Get("name")
	Println(r.Value)
	Println(opts.Get("missing").OK)
	// Output:
	// api
	// false
}

// ExampleOptions_Has checks for a value through `Options.Has` for agent options. Options
// carry loosely typed inputs while typed accessors keep call sites small.
func ExampleOptions_Has() {
	opts := NewOptions(Option{Key: "debug", Value: true})
	Println(opts.Has("debug"))
	// Output: true
}

// ExampleOptions_String renders `Options.String` as a stable string for agent options.
// Options carry loosely typed inputs while typed accessors keep call sites small.
func ExampleOptions_String() {
	opts := NewOptions(Option{Key: "name", Value: "api"})
	Println(opts.String("name"))
	// Output: api
}

// ExampleOptions_Int reads an integer through `Options.Int` for agent options. Options
// carry loosely typed inputs while typed accessors keep call sites small.
func ExampleOptions_Int() {
	opts := NewOptions(Option{Key: "port", Value: 8080})
	Println(opts.Int("port"))
	// Output: 8080
}

// ExampleOptions_Bool reads a boolean through `Options.Bool` for agent options. Options
// carry loosely typed inputs while typed accessors keep call sites small.
func ExampleOptions_Bool() {
	opts := NewOptions(Option{Key: "debug", Value: true})
	Println(opts.Bool("debug"))
	// Output: true
}

// ExampleOptions_Len counts entries through `Options.Len` for agent options. Options carry
// loosely typed inputs while typed accessors keep call sites small.
func ExampleOptions_Len() {
	opts := NewOptions(Option{Key: "name", Value: "api"})
	Println(opts.Len())
	// Output: 1
}

// ExampleOptions_Items returns all entries through `Options.Items` for agent options.
// Options carry loosely typed inputs while typed accessors keep call sites small.
func ExampleOptions_Items() {
	opts := NewOptions(Option{Key: "name", Value: "api"})
	items := opts.Items()
	items[0].Value = "worker"

	Println(opts.String("name"))
	Println(items[0].Value)
	// Output:
	// api
	// worker
}
