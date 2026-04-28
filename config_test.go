package core_test

import (
	. "dappco.re/go"
)

// --- Config ---

func TestConfig_SetGet_Good(t *T) {
	c := New()
	c.Config().Set("api_url", "https://api.lthn.ai")
	c.Config().Set("max_agents", 5)

	r := c.Config().Get("api_url")
	AssertTrue(t, r.OK)
	AssertEqual(t, "https://api.lthn.ai", r.Value)
}

func TestConfig_Get_Bad(t *T) {
	c := New()
	r := c.Config().Get("missing")
	AssertFalse(t, r.OK)
	AssertNil(t, r.Value)
}

func TestConfig_TypedAccessors_Good(t *T) {
	c := New()
	c.Config().Set("url", "https://lthn.ai")
	c.Config().Set("port", 8080)
	c.Config().Set("debug", true)

	AssertEqual(t, "https://lthn.ai", c.Config().String("url"))
	AssertEqual(t, 8080, c.Config().Int("port"))
	AssertTrue(t, c.Config().Bool("debug"))
}

func TestConfig_TypedAccessors_Bad(t *T) {
	c := New()
	// Missing keys return zero values
	AssertEqual(t, "", c.Config().String("missing"))
	AssertEqual(t, 0, c.Config().Int("missing"))
	AssertFalse(t, c.Config().Bool("missing"))
}

// --- Feature Flags ---

func TestConfig_Features_Good(t *T) {
	c := New()
	c.Config().Enable("dark-mode")
	c.Config().Enable("beta")

	AssertTrue(t, c.Config().Enabled("dark-mode"))
	AssertTrue(t, c.Config().Enabled("beta"))
	AssertFalse(t, c.Config().Enabled("missing"))
}

func TestConfig_Features_Disable_Good(t *T) {
	c := New()
	c.Config().Enable("feature")
	AssertTrue(t, c.Config().Enabled("feature"))

	c.Config().Disable("feature")
	AssertFalse(t, c.Config().Enabled("feature"))
}

func TestConfig_Features_CaseSensitive(t *T) {
	c := New()
	c.Config().Enable("Feature")
	AssertTrue(t, c.Config().Enabled("Feature"))
	AssertFalse(t, c.Config().Enabled("feature"))
}

func TestConfig_EnabledFeatures_Good(t *T) {
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

func TestConfig_ConfigVar_Good(t *T) {
	v := NewConfigVar("hello")
	AssertTrue(t, v.IsSet())
	AssertEqual(t, "hello", v.Get())

	v.Set("world")
	AssertEqual(t, "world", v.Get())

	v.Unset()
	AssertFalse(t, v.IsSet())
	AssertEqual(t, "", v.Get())
}

// --- AX-7 canonical triplets ---

func TestConfig_Config_New_Good(t *T) {
	cfg := (&Config{}).New()
	AssertNotNil(t, cfg)
	AssertFalse(t, cfg.Get("agent.host").OK)
	AssertEmpty(t, cfg.EnabledFeatures())
}

func TestConfig_Config_New_Bad(t *T) {
	cfg := (&Config{}).New()
	AssertEqual(t, "", cfg.String("missing"))
	AssertEqual(t, 0, cfg.Int("missing"))
	AssertFalse(t, cfg.Bool("missing"))
}

func TestConfig_Config_New_Ugly(t *T) {
	cfg := (&Config{}).New()
	cfg.Set("agent.host", "homelab.lthn.sh")
	cfg = cfg.New()
	AssertFalse(t, cfg.Get("agent.host").OK)
}

func TestConfig_Config_Set_Good(t *T) {
	cfg := (&Config{}).New()
	cfg.Set("agent.host", "homelab.lthn.sh")
	r := cfg.Get("agent.host")
	AssertTrue(t, r.OK)
	AssertEqual(t, "homelab.lthn.sh", r.Value)
}

func TestConfig_Config_Set_Bad(t *T) {
	cfg := (&Config{}).New()
	cfg.Set("", "root-setting")
	r := cfg.Get("")
	AssertTrue(t, r.OK)
	AssertEqual(t, "root-setting", r.Value)
}

func TestConfig_Config_Set_Ugly(t *T) {
	cfg := (&Config{}).New()
	cfg.Set("agent.optional", nil)
	r := cfg.Get("agent.optional")
	AssertTrue(t, r.OK)
	AssertNil(t, r.Value)
}

func TestConfig_Config_Get_Good(t *T) {
	cfg := (&Config{}).New()
	cfg.Set("agent.mode", "review")
	r := cfg.Get("agent.mode")
	AssertTrue(t, r.OK)
	AssertEqual(t, "review", r.Value)
}

func TestConfig_Config_Get_Bad(t *T) {
	cfg := (&Config{}).New()
	r := cfg.Get("agent.mode")
	AssertFalse(t, r.OK)
}

func TestConfig_Config_Get_Ugly(t *T) {
	var cfg Config
	r := cfg.Get("agent.mode")
	AssertFalse(t, r.OK)
}

func TestConfig_Config_String_Good(t *T) {
	cfg := (&Config{}).New()
	cfg.Set("agent.host", "homelab.lthn.sh")
	AssertEqual(t, "homelab.lthn.sh", cfg.String("agent.host"))
}

func TestConfig_Config_String_Bad(t *T) {
	cfg := (&Config{}).New()
	cfg.Set("agent.port", 9101)
	AssertEqual(t, "", cfg.String("agent.port"))
}

func TestConfig_Config_String_Ugly(t *T) {
	var cfg Config
	AssertEqual(t, "", cfg.String("agent.host"))
}

func TestConfig_Config_Int_Good(t *T) {
	cfg := (&Config{}).New()
	cfg.Set("agent.port", 9101)
	AssertEqual(t, 9101, cfg.Int("agent.port"))
}

func TestConfig_Config_Int_Bad(t *T) {
	cfg := (&Config{}).New()
	cfg.Set("agent.port", "9101")
	AssertEqual(t, 0, cfg.Int("agent.port"))
}

func TestConfig_Config_Int_Ugly(t *T) {
	var cfg Config
	AssertEqual(t, 0, cfg.Int("agent.port"))
}

func TestConfig_Config_Bool_Good(t *T) {
	cfg := (&Config{}).New()
	cfg.Set("agent.enabled", true)
	AssertTrue(t, cfg.Bool("agent.enabled"))
}

func TestConfig_Config_Bool_Bad(t *T) {
	cfg := (&Config{}).New()
	cfg.Set("agent.enabled", "true")
	AssertFalse(t, cfg.Bool("agent.enabled"))
}

func TestConfig_Config_Bool_Ugly(t *T) {
	var cfg Config
	AssertFalse(t, cfg.Bool("agent.enabled"))
}

func TestConfig_ConfigGet_Good(t *T) {
	cfg := (&Config{}).New()
	cfg.Set("session.token", "lethean-session")
	AssertEqual(t, "lethean-session", ConfigGet[string](cfg, "session.token"))
}

func TestConfig_ConfigGet_Bad(t *T) {
	cfg := (&Config{}).New()
	cfg.Set("session.ttl", "60")
	AssertEqual(t, 0, ConfigGet[int](cfg, "session.ttl"))
}

func TestConfig_ConfigGet_Ugly(t *T) {
	AssertPanics(t, func() {
		ConfigGet[string](nil, "session.token")
	})
}

func TestConfig_Config_Enable_Good(t *T) {
	cfg := (&Config{}).New()
	cfg.Enable("agent.dispatch")
	AssertTrue(t, cfg.Enabled("agent.dispatch"))
}

func TestConfig_Config_Enable_Bad(t *T) {
	cfg := (&Config{}).New()
	cfg.Enable("")
	AssertTrue(t, cfg.Enabled(""))
}

func TestConfig_Config_Enable_Ugly(t *T) {
	var cfg Config
	cfg.Enable("agent.dispatch")
	AssertTrue(t, cfg.Enabled("agent.dispatch"))
}

func TestConfig_Config_Disable_Good(t *T) {
	cfg := (&Config{}).New()
	cfg.Enable("agent.dispatch")
	cfg.Disable("agent.dispatch")
	AssertFalse(t, cfg.Enabled("agent.dispatch"))
}

func TestConfig_Config_Disable_Bad(t *T) {
	cfg := (&Config{}).New()
	cfg.Disable("missing.feature")
	AssertFalse(t, cfg.Enabled("missing.feature"))
}

func TestConfig_Config_Disable_Ugly(t *T) {
	var cfg Config
	cfg.Disable("")
	AssertFalse(t, cfg.Enabled(""))
}

func TestConfig_Config_Enabled_Good(t *T) {
	cfg := (&Config{}).New()
	cfg.Enable("agent.dispatch")
	AssertTrue(t, cfg.Enabled("agent.dispatch"))
}

func TestConfig_Config_Enabled_Bad(t *T) {
	cfg := (&Config{}).New()
	AssertFalse(t, cfg.Enabled("agent.dispatch"))
}

func TestConfig_Config_Enabled_Ugly(t *T) {
	var cfg Config
	AssertFalse(t, cfg.Enabled("agent.dispatch"))
}

func TestConfig_Config_EnabledFeatures_Good(t *T) {
	cfg := (&Config{}).New()
	cfg.Enable("agent.dispatch")
	cfg.Enable("agent.review")
	features := cfg.EnabledFeatures()
	AssertContains(t, features, "agent.dispatch")
	AssertContains(t, features, "agent.review")
}

func TestConfig_Config_EnabledFeatures_Bad(t *T) {
	cfg := (&Config{}).New()
	AssertEmpty(t, cfg.EnabledFeatures())
}

func TestConfig_Config_EnabledFeatures_Ugly(t *T) {
	cfg := (&Config{}).New()
	cfg.Enable("agent.dispatch")
	cfg.Disable("agent.review")
	features := cfg.EnabledFeatures()
	AssertContains(t, features, "agent.dispatch")
	AssertNotContains(t, features, "agent.review")
}

func TestConfig_NewConfigVar_Good(t *T) {
	v := NewConfigVar("homelab.lthn.sh")
	AssertTrue(t, v.IsSet())
	AssertEqual(t, "homelab.lthn.sh", v.Get())
}

func TestConfig_NewConfigVar_Bad(t *T) {
	v := NewConfigVar("")
	AssertTrue(t, v.IsSet())
	AssertEqual(t, "", v.Get())
}

func TestConfig_NewConfigVar_Ugly(t *T) {
	v := NewConfigVar[*App](nil)
	AssertTrue(t, v.IsSet())
	AssertNil(t, v.Get())
}

func TestConfig_ConfigVar_Get_Good(t *T) {
	v := NewConfigVar("codex")
	AssertEqual(t, "codex", v.Get())
}

func TestConfig_ConfigVar_Get_Bad(t *T) {
	var v ConfigVar[string]
	AssertEqual(t, "", v.Get())
}

func TestConfig_ConfigVar_Get_Ugly(t *T) {
	v := NewConfigVar("session-token")
	v.Unset()
	AssertEqual(t, "", v.Get())
}

func TestConfig_ConfigVar_Set_Good(t *T) {
	var v ConfigVar[string]
	v.Set("codex")
	AssertTrue(t, v.IsSet())
	AssertEqual(t, "codex", v.Get())
}

func TestConfig_ConfigVar_Set_Bad(t *T) {
	var v ConfigVar[int]
	v.Set(0)
	AssertTrue(t, v.IsSet())
	AssertEqual(t, 0, v.Get())
}

func TestConfig_ConfigVar_Set_Ugly(t *T) {
	var v ConfigVar[*App]
	v.Set(nil)
	AssertTrue(t, v.IsSet())
	AssertNil(t, v.Get())
}

func TestConfig_ConfigVar_IsSet_Good(t *T) {
	v := NewConfigVar(true)
	AssertTrue(t, v.IsSet())
}

func TestConfig_ConfigVar_IsSet_Bad(t *T) {
	var v ConfigVar[bool]
	AssertFalse(t, v.IsSet())
}

func TestConfig_ConfigVar_IsSet_Ugly(t *T) {
	v := NewConfigVar(false)
	AssertTrue(t, v.IsSet(), "set false differs from unset")
	v.Unset()
	AssertFalse(t, v.IsSet())
}

func TestConfig_ConfigVar_Unset_Good(t *T) {
	v := NewConfigVar("session-token")
	v.Unset()
	AssertFalse(t, v.IsSet())
	AssertEqual(t, "", v.Get())
}

func TestConfig_ConfigVar_Unset_Bad(t *T) {
	var v ConfigVar[string]
	v.Unset()
	AssertFalse(t, v.IsSet())
	AssertEqual(t, "", v.Get())
}

func TestConfig_ConfigVar_Unset_Ugly(t *T) {
	v := NewConfigVar(&App{Name: "agent"})
	v.Unset()
	AssertFalse(t, v.IsSet())
	AssertNil(t, v.Get())
}
