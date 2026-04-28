// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

func TestInfo_Env_OS_Good(t *T) {
	v := Env("OS")
	AssertNotEmpty(t, v)
	AssertContains(t, []string{"darwin", "linux", "windows"}, v)
}

func TestInfo_Env_ARCH_Good(t *T) {
	v := Env("ARCH")
	AssertNotEmpty(t, v)
	AssertContains(t, []string{"amd64", "arm64", "386"}, v)
}

func TestInfo_Env_GO_Good(t *T) {
	AssertTrue(t, HasPrefix(Env("GO"), "go"))
}

func TestInfo_Env_DS_Good(t *T) {
	ds := Env("DS")
	AssertContains(t, []string{"/", "\\"}, ds)
}

func TestInfo_Env_PS_Good(t *T) {
	ps := Env("PS")
	AssertContains(t, []string{":", ";"}, ps)
}

func TestInfo_Env_DIR_HOME_Good(t *T) {
	home := Env("DIR_HOME")
	AssertNotEmpty(t, home)
	AssertTrue(t, PathIsAbs(home), "DIR_HOME should be absolute")
}

func TestInfo_Env_DIR_TMP_Good(t *T) {
	AssertNotEmpty(t, Env("DIR_TMP"))
}

func TestInfo_Env_DIR_CONFIG_Good(t *T) {
	AssertNotEmpty(t, Env("DIR_CONFIG"))
}

func TestInfo_Env_DIR_CACHE_Good(t *T) {
	AssertNotEmpty(t, Env("DIR_CACHE"))
}

func TestInfo_Env_HOSTNAME_Good(t *T) {
	AssertNotEmpty(t, Env("HOSTNAME"))
}

func TestInfo_Env_USER_Good(t *T) {
	AssertNotEmpty(t, Env("USER"))
}

func TestInfo_Env_PID_Good(t *T) {
	AssertNotEmpty(t, Env("PID"))
}

func TestInfo_Env_NUM_CPU_Good(t *T) {
	AssertNotEmpty(t, Env("NUM_CPU"))
}

func TestInfo_Env_CORE_START_Good(t *T) {
	ts := Env("CORE_START")
	RequireNotEmpty(t, ts)
	r := TimeParse(TimeRFC3339, ts)
	AssertTrue(t, r.OK, "CORE_START should be valid RFC3339")
}

func TestInfo_Env_Bad_Unknown(t *T) {
	AssertEqual(t, "", Env("NOPE"))
}

func TestInfo_Env_Good_CoreInstance(t *T) {
	c := New()
	AssertEqual(t, Env("OS"), c.Env("OS"))
	AssertEqual(t, Env("DIR_HOME"), c.Env("DIR_HOME"))
}

func TestInfo_EnvKeys_Good(t *T) {
	keys := EnvKeys()
	AssertNotEmpty(t, keys)
	AssertContains(t, keys, "OS")
	AssertContains(t, keys, "DIR_HOME")
	AssertContains(t, keys, "CORE_START")
}

func TestInfo_Arch_Good(t *T) {
	AssertNotEmpty(t, Arch())
}

func TestInfo_Arch_Bad(t *T) {
	AssertEqual(t, Env("ARCH"), Arch())
}

func TestInfo_Arch_Ugly(t *T) {
	AssertEqual(t, Lower(Arch()), Arch())
}

func TestInfo_Env_Bad(t *T) {
	AssertEqual(t, "", Env("CORE_TEST_MISSING"))
}

func TestInfo_Env_Ugly(t *T) {
	t.Setenv("CORE_TEST_SESSION", "token")
	AssertEqual(t, "token", Env("CORE_TEST_SESSION"))
}

func TestInfo_EnvKeys_Bad(t *T) {
	AssertNotContains(t, EnvKeys(), "CORE_TEST_MISSING")
}

func TestInfo_EnvKeys_Ugly(t *T) {
	keys := EnvKeys()
	RequireNotEmpty(t, keys)
	keys[0] = "MUTATED"

	AssertNotContains(t, EnvKeys(), "MUTATED")
}

func TestInfo_GoVersion_Good(t *T) {
	AssertTrue(t, HasPrefix(GoVersion(), "go"))
}

func TestInfo_GoVersion_Bad(t *T) {
	AssertEqual(t, Env("GO"), GoVersion())
}

func TestInfo_GoVersion_Ugly(t *T) {
	AssertNotContains(t, GoVersion(), " ")
}

func TestInfo_NumCPU_Good(t *T) {
	AssertGreaterOrEqual(t, NumCPU(), 1)
}

func TestInfo_NumCPU_Bad(t *T) {
	AssertEqual(t, Env("NUM_CPU"), Itoa(NumCPU()))
}

func TestInfo_NumCPU_Ugly(t *T) {
	AssertNotPanics(t, func() {
		_ = NumCPU()
	})
}

func TestInfo_OS_Bad(t *T) {
	AssertEqual(t, Env("OS"), OS())
}

func TestInfo_OS_Ugly(t *T) {
	AssertEqual(t, Lower(OS()), OS())
}

func TestInfo_StackBuf_Good(t *T) {
	stack := string(StackBuf())
	AssertContains(t, stack, "goroutine")
}

func TestInfo_StackBuf_Bad(t *T) {
	AssertNotEmpty(t, StackBuf())
}

func TestInfo_StackBuf_Ugly(t *T) {
	AssertNotEqual(t, "", string(StackBuf()))
}
