// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"os"
	"strings"
	"testing"

	core "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnv_Setenv_Good(t *testing.T) {
	key := envTestKey(t, "GOOD")
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv(key))
	})

	require.NoError(t, core.Setenv(key, "ok"))
	assert.Equal(t, "ok", core.Env(key))
}

func TestEnv_Setenv_Bad(t *testing.T) {
	err := core.Setenv("CORE_GO_BAD=KEY", "bad")
	assert.Error(t, err)
}

func TestEnv_Setenv_Ugly(t *testing.T) {
	key := envTestKey(t, "UGLY")
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv(key))
	})

	require.NoError(t, core.Setenv(key, ""))
	value, ok := os.LookupEnv(key)
	assert.True(t, ok)
	assert.Equal(t, "", value)
}

func TestEnv_UnsetEnv_Good(t *testing.T) {
	key := envTestKey(t, "GOOD")
	require.NoError(t, os.Setenv(key, "set"))
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv(key))
	})

	require.NoError(t, core.UnsetEnv(key))
	_, ok := os.LookupEnv(key)
	assert.False(t, ok)
	assert.Equal(t, "", core.Env(key))
}

func TestEnv_UnsetEnv_Bad(t *testing.T) {
	key := envTestKey(t, "BAD")
	require.NoError(t, os.Setenv(key, "set"))
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv(key))
	})

	assert.NoError(t, core.UnsetEnv(key+"_WRONG"))
	assert.Equal(t, "set", core.Env(key))
}

func TestEnv_UnsetEnv_Ugly(t *testing.T) {
	key := envTestKey(t, "UGLY")
	require.NoError(t, os.Unsetenv(key))

	assert.NoError(t, core.UnsetEnv(key))
	_, ok := os.LookupEnv(key)
	assert.False(t, ok)
}

func envTestKey(t *testing.T, suffix string) string {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_", "=", "_").Replace(t.Name())
	return "CORE_GO_" + name + "_" + suffix
}
