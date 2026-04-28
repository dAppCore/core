package core_test

import (
	. "dappco.re/go/core"
)

func ExampleConfig_Set() {
	c := New()
	c.Config().Set("database.host", "localhost")
	c.Config().Set("database.port", 5432)

	Println(c.Config().String("database.host"))
	Println(c.Config().Int("database.port"))
	// Output:
	// localhost
	// 5432
}

func ExampleConfig_Enable() {
	c := New()
	c.Config().Enable("dark-mode")
	c.Config().Enable("beta-features")

	Println(c.Config().Enabled("dark-mode"))
	features := c.Config().EnabledFeatures()
	SliceSort(features)
	Println(features)
	// Output:
	// true
	// [beta-features dark-mode]
}

func ExampleConfigVar() {
	v := NewConfigVar(42)
	Println(v.Get(), v.IsSet())

	v.Unset()
	Println(v.Get(), v.IsSet())
	// Output:
	// 42 true
	// 0 false
}
