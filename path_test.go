// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"os"
	"testing"

	core "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPath_Relative(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	ds := core.Env("DS")
	assert.Equal(t, home+ds+"Code"+ds+".core", core.Path("Code", ".core"))
}

func TestPath_Absolute(t *testing.T) {
	assert.Equal(t, "/tmp/workspace", core.Path("/tmp", "workspace"))
}

func TestPath_Empty(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	assert.Equal(t, home, core.Path())
}

func TestPath_Cleans(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	assert.Equal(t, home+core.Env("DS")+"Code", core.Path("Code", "sub", ".."))
}

func TestPath_CleanDoubleSlash(t *testing.T) {
	assert.Equal(t, "/tmp/file", core.Path("/tmp//file"))
}

func TestPathBase(t *testing.T) {
	assert.Equal(t, "core", core.PathBase("/Users/snider/Code/core"))
	assert.Equal(t, "homelab", core.PathBase("deploy/to/homelab"))
}

func TestPathBase_Root(t *testing.T) {
	assert.Equal(t, "/", core.PathBase("/"))
}

func TestPathBase_Empty(t *testing.T) {
	assert.Equal(t, ".", core.PathBase(""))
}

func TestPathDir(t *testing.T) {
	assert.Equal(t, "/Users/snider/Code", core.PathDir("/Users/snider/Code/core"))
}

func TestPathDir_Root(t *testing.T) {
	assert.Equal(t, "/", core.PathDir("/file"))
}

func TestPathDir_NoDir(t *testing.T) {
	assert.Equal(t, ".", core.PathDir("file.go"))
}

func TestPathExt(t *testing.T) {
	assert.Equal(t, ".go", core.PathExt("main.go"))
	assert.Equal(t, "", core.PathExt("Makefile"))
	assert.Equal(t, ".gz", core.PathExt("archive.tar.gz"))
}

func TestPath_EnvConsistency(t *testing.T) {
	// Path() and Env("DIR_HOME") should agree
	assert.Equal(t, core.Env("DIR_HOME"), core.Path())
}
