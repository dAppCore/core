// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"testing"
	"time"

	core "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfo_Env_OS_Good(t *testing.T) {
	v := core.Env("OS")
	assert.NotEmpty(t, v)
	assert.Contains(t, []string{"darwin", "linux", "windows"}, v)
}

func TestInfo_Env_ARCH_Good(t *testing.T) {
	v := core.Env("ARCH")
	assert.NotEmpty(t, v)
	assert.Contains(t, []string{"amd64", "arm64", "386"}, v)
}

func TestInfo_Env_GO_Good(t *testing.T) {
	assert.True(t, core.HasPrefix(core.Env("GO"), "go"))
}

func TestInfo_Env_DS_Good(t *testing.T) {
	ds := core.Env("DS")
	assert.Contains(t, []string{"/", "\\"}, ds)
}

func TestInfo_Env_PS_Good(t *testing.T) {
	ps := core.Env("PS")
	assert.Contains(t, []string{":", ";"}, ps)
}

func TestInfo_Env_DIR_HOME_Good(t *testing.T) {
	home := core.Env("DIR_HOME")
	assert.NotEmpty(t, home)
	assert.True(t, core.PathIsAbs(home), "DIR_HOME should be absolute")
}

func TestInfo_Env_DIR_TMP_Good(t *testing.T) {
	assert.NotEmpty(t, core.Env("DIR_TMP"))
}

func TestInfo_Env_DIR_CONFIG_Good(t *testing.T) {
	assert.NotEmpty(t, core.Env("DIR_CONFIG"))
}

func TestInfo_Env_DIR_CACHE_Good(t *testing.T) {
	assert.NotEmpty(t, core.Env("DIR_CACHE"))
}

func TestInfo_Env_HOSTNAME_Good(t *testing.T) {
	assert.NotEmpty(t, core.Env("HOSTNAME"))
}

func TestInfo_Env_USER_Good(t *testing.T) {
	assert.NotEmpty(t, core.Env("USER"))
}

func TestInfo_Env_PID_Good(t *testing.T) {
	assert.NotEmpty(t, core.Env("PID"))
}

func TestInfo_Env_NUM_CPU_Good(t *testing.T) {
	assert.NotEmpty(t, core.Env("NUM_CPU"))
}

func TestInfo_Env_CORE_START_Good(t *testing.T) {
	ts := core.Env("CORE_START")
	require.NotEmpty(t, ts)
	_, err := time.Parse(time.RFC3339, ts)
	assert.NoError(t, err, "CORE_START should be valid RFC3339")
}

func TestInfo_Env_Bad_Unknown(t *testing.T) {
	assert.Equal(t, "", core.Env("NOPE"))
}

func TestInfo_Env_Good_CoreInstance(t *testing.T) {
	c := core.New()
	assert.Equal(t, core.Env("OS"), c.Env("OS"))
	assert.Equal(t, core.Env("DIR_HOME"), c.Env("DIR_HOME"))
}

func TestInfo_EnvKeys_Good(t *testing.T) {
	keys := core.EnvKeys()
	assert.NotEmpty(t, keys)
	assert.Contains(t, keys, "OS")
	assert.Contains(t, keys, "DIR_HOME")
	assert.Contains(t, keys, "CORE_START")
}
