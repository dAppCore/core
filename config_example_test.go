package core_test

import (
	"fmt"

	. "dappco.re/go/core"
)

func ExampleConfig_Set() {
	c := New()
	c.Config().Set("database.host", "localhost")
	c.Config().Set("database.port", 5432)

	fmt.Println(c.Config().String("database.host"))
	fmt.Println(c.Config().Int("database.port"))
	// Output:
	// localhost
	// 5432
}

func ExampleConfig_Enable() {
	c := New()
	c.Config().Enable("dark-mode")
	c.Config().Enable("beta-features")

	fmt.Println(c.Config().Enabled("dark-mode"))
	fmt.Println(c.Config().EnabledFeatures())
	// Output:
	// true
	// [dark-mode beta-features]
}

func ExampleConfigVar() {
	v := NewConfigVar(42)
	fmt.Println(v.Get(), v.IsSet())

	v.Unset()
	fmt.Println(v.Get(), v.IsSet())
	// Output:
	// 42 true
	// 0 false
}
