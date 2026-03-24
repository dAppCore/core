// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"os"
	"runtime"
	"testing"
	"time"

	core "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnv_OS(t *testing.T) {
	assert.Equal(t, runtime.GOOS, core.Env("OS"))
}

func TestEnv_ARCH(t *testing.T) {
	assert.Equal(t, runtime.GOARCH, core.Env("ARCH"))
}

func TestEnv_GO(t *testing.T) {
	assert.Equal(t, runtime.Version(), core.Env("GO"))
}

func TestEnv_DS(t *testing.T) {
	assert.Equal(t, string(os.PathSeparator), core.Env("DS"))
}

func TestEnv_PS(t *testing.T) {
	assert.Equal(t, string(os.PathListSeparator), core.Env("PS"))
}

func TestEnv_DIR_HOME(t *testing.T) {
	if ch := os.Getenv("CORE_HOME"); ch != "" {
		assert.Equal(t, ch, core.Env("DIR_HOME"))
		return
	}
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	assert.Equal(t, home, core.Env("DIR_HOME"))
}

func TestEnv_DIR_TMP(t *testing.T) {
	assert.Equal(t, os.TempDir(), core.Env("DIR_TMP"))
}

func TestEnv_DIR_CONFIG(t *testing.T) {
	cfg, err := os.UserConfigDir()
	require.NoError(t, err)
	assert.Equal(t, cfg, core.Env("DIR_CONFIG"))
}

func TestEnv_DIR_CACHE(t *testing.T) {
	cache, err := os.UserCacheDir()
	require.NoError(t, err)
	assert.Equal(t, cache, core.Env("DIR_CACHE"))
}

func TestEnv_HOSTNAME(t *testing.T) {
	hostname, err := os.Hostname()
	require.NoError(t, err)
	assert.Equal(t, hostname, core.Env("HOSTNAME"))
}

func TestEnv_USER(t *testing.T) {
	assert.NotEmpty(t, core.Env("USER"))
}

func TestEnv_PID(t *testing.T) {
	assert.NotEmpty(t, core.Env("PID"))
}

func TestEnv_NUM_CPU(t *testing.T) {
	assert.NotEmpty(t, core.Env("NUM_CPU"))
}

func TestEnv_CORE_START(t *testing.T) {
	ts := core.Env("CORE_START")
	require.NotEmpty(t, ts)
	_, err := time.Parse(time.RFC3339, ts)
	assert.NoError(t, err, "CORE_START should be valid RFC3339")
}

func TestEnv_Unknown(t *testing.T) {
	assert.Equal(t, "", core.Env("NOPE"))
}

func TestEnv_CoreInstance(t *testing.T) {
	c := core.New().Value.(*core.Core)
	assert.Equal(t, core.Env("OS"), c.Env("OS"))
	assert.Equal(t, core.Env("DIR_HOME"), c.Env("DIR_HOME"))
}

func TestEnvKeys(t *testing.T) {
	keys := core.EnvKeys()
	assert.NotEmpty(t, keys)
	assert.Contains(t, keys, "OS")
	assert.Contains(t, keys, "DIR_HOME")
	assert.Contains(t, keys, "CORE_START")
}
