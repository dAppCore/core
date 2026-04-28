package core_test

import (
	. "dappco.re/go"
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

func ExampleNewConfigVar_config() {
	v := NewConfigVar("enabled")
	Println(v.Get())
	Println(v.IsSet())
	// Output:
	// enabled
	// true
}

func ExampleConfigVar_Set() {
	var v ConfigVar[string]
	v.Set("enabled")
	Println(v.Get())
	Println(v.IsSet())
	// Output:
	// enabled
	// true
}

func ExampleConfigVar_Unset() {
	v := NewConfigVar("enabled")
	v.Unset()
	Println(v.Get())
	Println(v.IsSet())
	// Output:
	//
	// false
}

func ExampleConfig_New() {
	cfg := (&Config{}).New()
	Println(cfg.EnabledFeatures())
	// Output: []
}

func ExampleConfigOptions() {
	opts := ConfigOptions{
		Settings: map[string]any{"host": "localhost"},
		Features: map[string]bool{"debug": true},
	}
	Println(opts.Settings["host"])
	Println(opts.Features["debug"])
	// Output:
	// localhost
	// true
}

func ExampleConfig_Get() {
	cfg := (&Config{}).New()
	cfg.Set("host", "localhost")
	Println(cfg.Get("host").Value)
	Println(cfg.Get("missing").OK)
	// Output:
	// localhost
	// false
}

func ExampleConfig_String() {
	cfg := (&Config{}).New()
	cfg.Set("host", "localhost")
	Println(cfg.String("host"))
	// Output: localhost
}

func ExampleConfig_Int() {
	cfg := (&Config{}).New()
	cfg.Set("port", 8080)
	Println(cfg.Int("port"))
	// Output: 8080
}

func ExampleConfig_Bool() {
	cfg := (&Config{}).New()
	cfg.Set("debug", true)
	Println(cfg.Bool("debug"))
	// Output: true
}

func ExampleConfigGet() {
	cfg := (&Config{}).New()
	cfg.Set("host", "localhost")
	Println(ConfigGet[string](cfg, "host"))
	// Output: localhost
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

func ExampleConfig_Disable() {
	c := New()
	c.Config().Enable("debug")
	c.Config().Disable("debug")
	Println(c.Config().Enabled("debug"))
	// Output: false
}

func ExampleConfig_Enabled() {
	c := New()
	c.Config().Enable("debug")
	Println(c.Config().Enabled("debug"))
	// Output: true
}

func ExampleConfig_EnabledFeatures() {
	c := New()
	c.Config().Enable("beta")
	c.Config().Enable("debug")
	features := c.Config().EnabledFeatures()
	SliceSort(features)
	Println(features)
	// Output: [beta debug]
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
