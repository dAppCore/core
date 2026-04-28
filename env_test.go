// SPDX-License-Identifier: EUPL-1.2

package core_test

import . "dappco.re/go"

func TestEnv_Setenv_Good(t *T) {
	key := envTestKey(t, "GOOD")
	t.Cleanup(func() {
		RequireTrue(t, Unsetenv(key).OK)
	})

	RequireTrue(t, Setenv(key, "ok").OK)
	AssertEqual(t, "ok", Env(key))
}

func TestEnv_Setenv_Bad(t *T) {
	r := Setenv("CORE_GO_BAD=KEY", "bad")
	AssertFalse(t, r.OK)
}

func TestEnv_Setenv_Ugly(t *T) {
	key := envTestKey(t, "UGLY")
	t.Cleanup(func() {
		RequireTrue(t, Unsetenv(key).OK)
	})

	RequireTrue(t, Setenv(key, "").OK)
	value, ok := LookupEnv(key)
	AssertTrue(t, ok)
	AssertEqual(t, "", value)
}

func TestEnv_Unsetenv_Good(t *T) {
	key := envTestKey(t, "GOOD")
	RequireTrue(t, Setenv(key, "set").OK)
	t.Cleanup(func() {
		RequireTrue(t, Unsetenv(key).OK)
	})

	RequireTrue(t, Unsetenv(key).OK)
	_, ok := LookupEnv(key)
	AssertFalse(t, ok)
	AssertEqual(t, "", Env(key))
}

func TestEnv_Unsetenv_Bad(t *T) {
	key := envTestKey(t, "BAD")
	RequireTrue(t, Setenv(key, "set").OK)
	t.Cleanup(func() {
		RequireTrue(t, Unsetenv(key).OK)
	})

	// Unsetting a never-set key still succeeds (idempotent).
	AssertTrue(t, Unsetenv(key+"_WRONG").OK)
	AssertEqual(t, "set", Env(key))
}

func TestEnv_Unsetenv_Ugly(t *T) {
	key := envTestKey(t, "UGLY")
	RequireTrue(t, Unsetenv(key).OK)

	AssertTrue(t, Unsetenv(key).OK)
	_, ok := LookupEnv(key)
	AssertFalse(t, ok)
}

func envTestKey(t *T, suffix string) string {
	t.Helper()

	name := Replace(Replace(Replace(t.Name(), "/", "_"), " ", "_"), "=", "_")
	return "CORE_GO_" + name + "_" + suffix
}
