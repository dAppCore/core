// SPDX-License-Identifier: EUPL-1.2

package core

func TestPath_lastIndex_Good(t *T) {
	AssertEqual(t, 9, lastIndex("deploy/to/homelab", "/"))
}
func TestPath_lastIndex_Bad(t *T) {
	AssertEqual(t, -1, lastIndex("deploy/to/homelab", "."))
}
func TestPath_lastIndex_Ugly(t *T) {
	AssertEqual(t, -1, lastIndex("deploy/to/homelab", ""))
}
