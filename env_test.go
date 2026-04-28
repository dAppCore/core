// SPDX-License-Identifier: EUPL-1.2

package core_test

import . "dappco.re/go"

func TestEnv_Setenv_Good(t *T) {
	key := envTestKey(t, "GOOD")
	t.Cleanup(func() {
		RequireNoError(t, Unsetenv(key))
	})

	RequireNoError(t, Setenv(key, "ok"))
	AssertEqual(t, "ok", Env(key))
}

func TestEnv_Setenv_Bad(t *T) {
	err := Setenv("CORE_GO_BAD=KEY", "bad")
	AssertError(t, err)
}

func TestEnv_Setenv_Ugly(t *T) {
	key := envTestKey(t, "UGLY")
	t.Cleanup(func() {
		RequireNoError(t, Unsetenv(key))
	})

	RequireNoError(t, Setenv(key, ""))
	value, ok := LookupEnv(key)
	AssertTrue(t, ok)
	AssertEqual(t, "", value)
}

func TestEnv_Unsetenv_Good(t *T) {
	key := envTestKey(t, "GOOD")
	RequireNoError(t, Setenv(key, "set"))
	t.Cleanup(func() {
		RequireNoError(t, Unsetenv(key))
	})

	RequireNoError(t, Unsetenv(key))
	_, ok := LookupEnv(key)
	AssertFalse(t, ok)
	AssertEqual(t, "", Env(key))
}

func TestEnv_Unsetenv_Bad(t *T) {
	key := envTestKey(t, "BAD")
	RequireNoError(t, Setenv(key, "set"))
	t.Cleanup(func() {
		RequireNoError(t, Unsetenv(key))
	})

	AssertNoError(t, Unsetenv(key+"_WRONG"))
	AssertEqual(t, "set", Env(key))
}

func TestEnv_Unsetenv_Ugly(t *T) {
	key := envTestKey(t, "UGLY")
	RequireNoError(t, Unsetenv(key))

	AssertNoError(t, Unsetenv(key))
	_, ok := LookupEnv(key)
	AssertFalse(t, ok)
}

func envTestKey(t *T, suffix string) string {
	t.Helper()

	name := Replace(Replace(Replace(t.Name(), "/", "_"), " ", "_"), "=", "_")
	return "CORE_GO_" + name + "_" + suffix
}
