// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"testing"
	"time"

	. "dappco.re/go/core"
)

func TestInfo_Env_OS_Good(t *testing.T) {
	v := Env("OS")
	AssertNotEmpty(t, v)
	AssertContains(t, []string{"darwin", "linux", "windows"}, v)
}

func TestInfo_Env_ARCH_Good(t *testing.T) {
	v := Env("ARCH")
	AssertNotEmpty(t, v)
	AssertContains(t, []string{"amd64", "arm64", "386"}, v)
}

func TestInfo_Env_GO_Good(t *testing.T) {
	AssertTrue(t, HasPrefix(Env("GO"), "go"))
}

func TestInfo_Env_DS_Good(t *testing.T) {
	ds := Env("DS")
	AssertContains(t, []string{"/", "\\"}, ds)
}

func TestInfo_Env_PS_Good(t *testing.T) {
	ps := Env("PS")
	AssertContains(t, []string{":", ";"}, ps)
}

func TestInfo_Env_DIR_HOME_Good(t *testing.T) {
	home := Env("DIR_HOME")
	AssertNotEmpty(t, home)
	AssertTrue(t, PathIsAbs(home), "DIR_HOME should be absolute")
}

func TestInfo_Env_DIR_TMP_Good(t *testing.T) {
	AssertNotEmpty(t, Env("DIR_TMP"))
}

func TestInfo_Env_DIR_CONFIG_Good(t *testing.T) {
	AssertNotEmpty(t, Env("DIR_CONFIG"))
}

func TestInfo_Env_DIR_CACHE_Good(t *testing.T) {
	AssertNotEmpty(t, Env("DIR_CACHE"))
}

func TestInfo_Env_HOSTNAME_Good(t *testing.T) {
	AssertNotEmpty(t, Env("HOSTNAME"))
}

func TestInfo_Env_USER_Good(t *testing.T) {
	AssertNotEmpty(t, Env("USER"))
}

func TestInfo_Env_PID_Good(t *testing.T) {
	AssertNotEmpty(t, Env("PID"))
}

func TestInfo_Env_NUM_CPU_Good(t *testing.T) {
	AssertNotEmpty(t, Env("NUM_CPU"))
}

func TestInfo_Env_CORE_START_Good(t *testing.T) {
	ts := Env("CORE_START")
	RequireNotEmpty(t, ts)
	_, err := time.Parse(time.RFC3339, ts)
	AssertNoError(t, err, "CORE_START should be valid RFC3339")
}

func TestInfo_Env_Bad_Unknown(t *testing.T) {
	AssertEqual(t, "", Env("NOPE"))
}

func TestInfo_Env_Good_CoreInstance(t *testing.T) {
	c := New()
	AssertEqual(t, Env("OS"), c.Env("OS"))
	AssertEqual(t, Env("DIR_HOME"), c.Env("DIR_HOME"))
}

func TestInfo_EnvKeys_Good(t *testing.T) {
	keys := EnvKeys()
	AssertNotEmpty(t, keys)
	AssertContains(t, keys, "OS")
	AssertContains(t, keys, "DIR_HOME")
	AssertContains(t, keys, "CORE_START")
}
