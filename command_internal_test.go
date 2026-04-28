// SPDX-License-Identifier: EUPL-1.2

package core

func TestCommand_pathName_Good(t *T) {
	AssertEqual(t, "homelab", pathName("deploy/to/homelab"))
}
func TestCommand_pathName_Bad(t *T) {
	AssertEqual(t, "", pathName(""))
}
func TestCommand_pathName_Ugly(t *T) {
	AssertEqual(t, "", pathName("deploy/"))
}
