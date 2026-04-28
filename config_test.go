package core_test

import (
	"testing"

	. "dappco.re/go/core"
)

// --- Config ---

func TestConfig_SetGet_Good(t *testing.T) {
	c := New()
	c.Config().Set("api_url", "https://api.lthn.ai")
	c.Config().Set("max_agents", 5)

	r := c.Config().Get("api_url")
	AssertTrue(t, r.OK)
	AssertEqual(t, "https://api.lthn.ai", r.Value)
}

func TestConfig_Get_Bad(t *testing.T) {
	c := New()
	r := c.Config().Get("missing")
	AssertFalse(t, r.OK)
	AssertNil(t, r.Value)
}

func TestConfig_TypedAccessors_Good(t *testing.T) {
	c := New()
	c.Config().Set("url", "https://lthn.ai")
	c.Config().Set("port", 8080)
	c.Config().Set("debug", true)

	AssertEqual(t, "https://lthn.ai", c.Config().String("url"))
	AssertEqual(t, 8080, c.Config().Int("port"))
	AssertTrue(t, c.Config().Bool("debug"))
}

func TestConfig_TypedAccessors_Bad(t *testing.T) {
	c := New()
	// Missing keys return zero values
	AssertEqual(t, "", c.Config().String("missing"))
	AssertEqual(t, 0, c.Config().Int("missing"))
	AssertFalse(t, c.Config().Bool("missing"))
}

// --- Feature Flags ---

func TestConfig_Features_Good(t *testing.T) {
	c := New()
	c.Config().Enable("dark-mode")
	c.Config().Enable("beta")

	AssertTrue(t, c.Config().Enabled("dark-mode"))
	AssertTrue(t, c.Config().Enabled("beta"))
	AssertFalse(t, c.Config().Enabled("missing"))
}

func TestConfig_Features_Disable_Good(t *testing.T) {
	c := New()
	c.Config().Enable("feature")
	AssertTrue(t, c.Config().Enabled("feature"))

	c.Config().Disable("feature")
	AssertFalse(t, c.Config().Enabled("feature"))
}

func TestConfig_Features_CaseSensitive(t *testing.T) {
	c := New()
	c.Config().Enable("Feature")
	AssertTrue(t, c.Config().Enabled("Feature"))
	AssertFalse(t, c.Config().Enabled("feature"))
}

func TestConfig_EnabledFeatures_Good(t *testing.T) {
	c := New()
	c.Config().Enable("a")
	c.Config().Enable("b")
	c.Config().Enable("c")
	c.Config().Disable("b")

	features := c.Config().EnabledFeatures()
	AssertContains(t, features, "a")
	AssertContains(t, features, "c")
	AssertNotContains(t, features, "b")
}

// --- ConfigVar ---

func TestConfig_ConfigVar_Good(t *testing.T) {
	v := NewConfigVar("hello")
	AssertTrue(t, v.IsSet())
	AssertEqual(t, "hello", v.Get())

	v.Set("world")
	AssertEqual(t, "world", v.Get())

	v.Unset()
	AssertFalse(t, v.IsSet())
	AssertEqual(t, "", v.Get())
}
