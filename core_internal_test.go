// SPDX-License-Identifier: EUPL-1.2

package core

func TestCore_registryProxy_Good(t *T) {
	src := NewRegistry[string]()
	src.Set("agent", "codex")

	proxy := registryProxy(src)

	AssertEqual(t, []string{"agent"}, proxy.Names())
	AssertEqual(t, "codex", proxy.Get("agent").Value)
}
func TestCore_registryProxy_Bad(t *T) {
	src := NewRegistry[string]()
	src.Set("agent", "codex")
	src.Disable("agent")

	proxy := registryProxy(src)

	AssertEqual(t, 0, proxy.Len())
}
func TestCore_registryProxy_Ugly(t *T) {
	AssertPanics(t, func() {
		registryProxy[int](nil)
	})
}
