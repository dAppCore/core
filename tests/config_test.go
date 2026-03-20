package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- Config ---

func TestConfig_SetGet_Good(t *testing.T) {
	c := New()
	c.Config().Set("api_url", "https://api.lthn.ai")
	c.Config().Set("max_agents", 5)

	val, ok := c.Config().Get("api_url")
	assert.True(t, ok)
	assert.Equal(t, "https://api.lthn.ai", val)
}

func TestConfig_Get_Bad(t *testing.T) {
	c := New()
	val, ok := c.Config().Get("missing")
	assert.False(t, ok)
	assert.Nil(t, val)
}

func TestConfig_TypedAccessors_Good(t *testing.T) {
	c := New()
	c.Config().Set("url", "https://lthn.ai")
	c.Config().Set("port", 8080)
	c.Config().Set("debug", true)

	assert.Equal(t, "https://lthn.ai", c.Config().String("url"))
	assert.Equal(t, 8080, c.Config().Int("port"))
	assert.True(t, c.Config().Bool("debug"))
}

func TestConfig_TypedAccessors_Bad(t *testing.T) {
	c := New()
	// Missing keys return zero values
	assert.Equal(t, "", c.Config().String("missing"))
	assert.Equal(t, 0, c.Config().Int("missing"))
	assert.False(t, c.Config().Bool("missing"))
}

// --- Feature Flags ---

func TestConfig_Features_Good(t *testing.T) {
	c := New()
	c.Config().Enable("dark-mode")
	c.Config().Enable("beta")

	assert.True(t, c.Config().Enabled("dark-mode"))
	assert.True(t, c.Config().Enabled("beta"))
	assert.False(t, c.Config().Enabled("missing"))
}

func TestConfig_Features_Disable_Good(t *testing.T) {
	c := New()
	c.Config().Enable("feature")
	assert.True(t, c.Config().Enabled("feature"))

	c.Config().Disable("feature")
	assert.False(t, c.Config().Enabled("feature"))
}

func TestConfig_Features_CaseSensitive(t *testing.T) {
	c := New()
	c.Config().Enable("Feature")
	assert.True(t, c.Config().Enabled("Feature"))
	assert.False(t, c.Config().Enabled("feature"))
}

func TestConfig_EnabledFeatures_Good(t *testing.T) {
	c := New()
	c.Config().Enable("a")
	c.Config().Enable("b")
	c.Config().Enable("c")
	c.Config().Disable("b")

	features := c.Config().EnabledFeatures()
	assert.Contains(t, features, "a")
	assert.Contains(t, features, "c")
	assert.NotContains(t, features, "b")
}

// --- ConfigVar ---

func TestConfigVar_Good(t *testing.T) {
	v := NewConfigVar("hello")
	assert.True(t, v.IsSet())
	assert.Equal(t, "hello", v.Get())

	v.Set("world")
	assert.Equal(t, "world", v.Get())

	v.Unset()
	assert.False(t, v.IsSet())
	assert.Equal(t, "", v.Get())
}
