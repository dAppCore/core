// SPDX-License-Identifier: EUPL-1.2

package core

func TestConfig_ConfigOptions_init_Good(t *T) {
	var opts ConfigOptions

	opts.init()

	AssertNotNil(t, opts.Settings)
	AssertNotNil(t, opts.Features)
}
func TestConfig_ConfigOptions_init_Bad(t *T) {
	opts := ConfigOptions{Settings: map[string]any{"agent": "codex"}}

	opts.init()

	AssertEqual(t, "codex", opts.Settings["agent"])
	AssertNotNil(t, opts.Features)
}
func TestConfig_ConfigOptions_init_Ugly(t *T) {
	opts := ConfigOptions{}

	opts.init()
	opts.Settings["site"] = "homelab"
	opts.init()

	AssertEqual(t, "homelab", opts.Settings["site"])
}
