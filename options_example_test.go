package core_test

import . "dappco.re/go"

func ExampleResult_New() {
	r := Result{}.New("created", nil)
	Println(r.OK)
	Println(r.Value)
	// Output:
	// true
	// created
}

func ExampleOption() {
	opt := Option{Key: "port", Value: 8080}
	Println(opt.Key)
	Println(opt.Value)
	// Output:
	// port
	// 8080
}

func ExampleNewOptions_withItems() {
	opts := NewOptions(
		Option{Key: "name", Value: "api"},
		Option{Key: "port", Value: 8080},
	)
	Println(opts.Len())
	// Output: 2
}

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

func ExampleOptions_Get() {
	opts := NewOptions(Option{Key: "name", Value: "api"})
	r := opts.Get("name")
	Println(r.Value)
	Println(opts.Get("missing").OK)
	// Output:
	// api
	// false
}

func ExampleOptions_Has() {
	opts := NewOptions(Option{Key: "debug", Value: true})
	Println(opts.Has("debug"))
	// Output: true
}

func ExampleOptions_String() {
	opts := NewOptions(Option{Key: "name", Value: "api"})
	Println(opts.String("name"))
	// Output: api
}

func ExampleOptions_Int() {
	opts := NewOptions(Option{Key: "port", Value: 8080})
	Println(opts.Int("port"))
	// Output: 8080
}

func ExampleOptions_Bool() {
	opts := NewOptions(Option{Key: "debug", Value: true})
	Println(opts.Bool("debug"))
	// Output: true
}

func ExampleOptions_Len() {
	opts := NewOptions(Option{Key: "name", Value: "api"})
	Println(opts.Len())
	// Output: 1
}

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
