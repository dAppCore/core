// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"os"
	"strings"
	"testing"

	. "dappco.re/go/core"
)

func TestEnv_Setenv_Good(t *testing.T) {
	key := envTestKey(t, "GOOD")
	t.Cleanup(func() {
		RequireNoError(t, os.Unsetenv(key))
	})

	RequireNoError(t, Setenv(key, "ok"))
	AssertEqual(t, "ok", Env(key))
}

func TestEnv_Setenv_Bad(t *testing.T) {
	err := Setenv("CORE_GO_BAD=KEY", "bad")
	AssertError(t, err)
}

func TestEnv_Setenv_Ugly(t *testing.T) {
	key := envTestKey(t, "UGLY")
	t.Cleanup(func() {
		RequireNoError(t, os.Unsetenv(key))
	})

	RequireNoError(t, Setenv(key, ""))
	value, ok := os.LookupEnv(key)
	AssertTrue(t, ok)
	AssertEqual(t, "", value)
}

func TestEnv_UnsetEnv_Good(t *testing.T) {
	key := envTestKey(t, "GOOD")
	RequireNoError(t, os.Setenv(key, "set"))
	t.Cleanup(func() {
		RequireNoError(t, os.Unsetenv(key))
	})

	RequireNoError(t, UnsetEnv(key))
	_, ok := os.LookupEnv(key)
	AssertFalse(t, ok)
	AssertEqual(t, "", Env(key))
}

func TestEnv_UnsetEnv_Bad(t *testing.T) {
	key := envTestKey(t, "BAD")
	RequireNoError(t, os.Setenv(key, "set"))
	t.Cleanup(func() {
		RequireNoError(t, os.Unsetenv(key))
	})

	AssertNoError(t, UnsetEnv(key+"_WRONG"))
	AssertEqual(t, "set", Env(key))
}

func TestEnv_UnsetEnv_Ugly(t *testing.T) {
	key := envTestKey(t, "UGLY")
	RequireNoError(t, os.Unsetenv(key))

	AssertNoError(t, UnsetEnv(key))
	_, ok := os.LookupEnv(key)
	AssertFalse(t, ok)
}

func envTestKey(t *testing.T, suffix string) string {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_", "=", "_").Replace(t.Name())
	return "CORE_GO_" + name + "_" + suffix
}
