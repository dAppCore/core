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

func TestInfo_Env_OS(t *testing.T) {
	assert.Equal(t, runtime.GOOS, core.Env("OS"))
}

func TestInfo_Env_ARCH(t *testing.T) {
	assert.Equal(t, runtime.GOARCH, core.Env("ARCH"))
}

func TestInfo_Env_GO(t *testing.T) {
	assert.Equal(t, runtime.Version(), core.Env("GO"))
}

func TestInfo_Env_DS(t *testing.T) {
	assert.Equal(t, string(os.PathSeparator), core.Env("DS"))
}

func TestInfo_Env_PS(t *testing.T) {
	assert.Equal(t, string(os.PathListSeparator), core.Env("PS"))
}

func TestInfo_Env_DIR_HOME(t *testing.T) {
	if ch := os.Getenv("CORE_HOME"); ch != "" {
		assert.Equal(t, ch, core.Env("DIR_HOME"))
		return
	}
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	assert.Equal(t, home, core.Env("DIR_HOME"))
}

func TestInfo_Env_DIR_TMP(t *testing.T) {
	assert.Equal(t, os.TempDir(), core.Env("DIR_TMP"))
}

func TestInfo_Env_DIR_CONFIG(t *testing.T) {
	cfg, err := os.UserConfigDir()
	require.NoError(t, err)
	assert.Equal(t, cfg, core.Env("DIR_CONFIG"))
}

func TestInfo_Env_DIR_CACHE(t *testing.T) {
	cache, err := os.UserCacheDir()
	require.NoError(t, err)
	assert.Equal(t, cache, core.Env("DIR_CACHE"))
}

func TestInfo_Env_HOSTNAME(t *testing.T) {
	hostname, err := os.Hostname()
	require.NoError(t, err)
	assert.Equal(t, hostname, core.Env("HOSTNAME"))
}

func TestInfo_Env_USER(t *testing.T) {
	assert.NotEmpty(t, core.Env("USER"))
}

func TestInfo_Env_PID(t *testing.T) {
	assert.NotEmpty(t, core.Env("PID"))
}

func TestInfo_Env_NUM_CPU(t *testing.T) {
	assert.NotEmpty(t, core.Env("NUM_CPU"))
}

func TestInfo_Env_CORE_START(t *testing.T) {
	ts := core.Env("CORE_START")
	require.NotEmpty(t, ts)
	_, err := time.Parse(time.RFC3339, ts)
	assert.NoError(t, err, "CORE_START should be valid RFC3339")
}

func TestInfo_Env_Unknown(t *testing.T) {
	assert.Equal(t, "", core.Env("NOPE"))
}

func TestInfo_Env_CoreInstance(t *testing.T) {
	c := core.New()
	assert.Equal(t, core.Env("OS"), c.Env("OS"))
	assert.Equal(t, core.Env("DIR_HOME"), c.Env("DIR_HOME"))
}

func TestInfo_EnvKeys(t *testing.T) {
	keys := core.EnvKeys()
	assert.NotEmpty(t, keys)
	assert.Contains(t, keys, "OS")
	assert.Contains(t, keys, "DIR_HOME")
	assert.Contains(t, keys, "CORE_START")
}
