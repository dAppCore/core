// SPDX-License-Identifier: EUPL-1.2

package core

func TestInfo_init_Good(t *T) {
	AssertNotEmpty(t, Env("OS"))
	AssertNotEmpty(t, Env("ARCH"))
	AssertNotEmpty(t, Env("DIR_HOME"))
}
func TestInfo_init_Bad(t *T) {
	AssertNotEmpty(t, Env("DIR_HOME"))
}
func TestInfo_init_Ugly(t *T) {
	AssertNotEmpty(t, Env("DIR_DATA"))
}
