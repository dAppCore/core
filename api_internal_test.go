// SPDX-License-Identifier: EUPL-1.2

package core

func TestApi_extractScheme_Good(t *T) {
	AssertEqual(t, "https", extractScheme("https://homelab.lthn.sh/api"))
}
func TestApi_extractScheme_Bad(t *T) {
	AssertEqual(t, "", extractScheme(""))
}
func TestApi_extractScheme_Ugly(t *T) {
	AssertEqual(t, "stdio", extractScheme("stdio"))
}
