// SPDX-License-Identifier: EUPL-1.2

package mnt_test

import (
	"embed"
	"os"
	"testing"

	"forge.lthn.ai/core/go/pkg/mnt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata
var testFS embed.FS

func TestFS_Good(t *testing.T) {
	sub, err := mnt.FS(testFS, "testdata")
	require.NoError(t, err)
	assert.Equal(t, "testdata", sub.BaseDir())
}

func TestFS_Bad_InvalidDir(t *testing.T) {
	_, err := mnt.FS(testFS, "nonexistent")
	assert.Error(t, err)
}

func TestSub_ReadFile_Good(t *testing.T) {
	sub, err := mnt.FS(testFS, "testdata")
	require.NoError(t, err)

	data, err := sub.ReadFile("hello.txt")
	require.NoError(t, err)
	assert.Equal(t, "hello world\n", string(data))
}

func TestSub_ReadString_Good(t *testing.T) {
	sub, err := mnt.FS(testFS, "testdata")
	require.NoError(t, err)

	content, err := sub.ReadString("hello.txt")
	require.NoError(t, err)
	assert.Equal(t, "hello world\n", content)
}

func TestSub_ReadDir_Good(t *testing.T) {
	sub, err := mnt.FS(testFS, "testdata")
	require.NoError(t, err)

	entries, err := sub.ReadDir(".")
	require.NoError(t, err)
	assert.True(t, len(entries) >= 1)
}

func TestSub_Sub_Good(t *testing.T) {
	sub, err := mnt.FS(testFS, "testdata")
	require.NoError(t, err)

	nested, err := sub.Sub("subdir")
	require.NoError(t, err)

	content, err := nested.ReadString("nested.txt")
	require.NoError(t, err)
	assert.Equal(t, "nested content\n", content)
}

func TestExtract_Good(t *testing.T) {
	sub, err := mnt.FS(testFS, "testdata/template")
	require.NoError(t, err)

	targetDir := t.TempDir()
	err = mnt.Extract(sub, targetDir, map[string]string{
		"Name":   "myproject",
		"Module": "forge.lthn.ai/core/myproject",
	})
	require.NoError(t, err)

	// Check template was processed
	readme, err := os.ReadFile(targetDir + "/README.md")
	require.NoError(t, err)
	assert.Contains(t, string(readme), "myproject project")

	// Check go.mod template was processed
	gomod, err := os.ReadFile(targetDir + "/go.mod")
	require.NoError(t, err)
	assert.Contains(t, string(gomod), "module forge.lthn.ai/core/myproject")

	// Check non-template was copied as-is
	main, err := os.ReadFile(targetDir + "/main.go")
	require.NoError(t, err)
	assert.Equal(t, "package main\n", string(main))
}

func TestExtract_Good_NoData(t *testing.T) {
	sub, err := mnt.FS(testFS, "testdata")
	require.NoError(t, err)

	targetDir := t.TempDir()
	err = mnt.Extract(sub, targetDir, nil)
	require.NoError(t, err)

	data, err := os.ReadFile(targetDir + "/hello.txt")
	require.NoError(t, err)
	assert.Equal(t, "hello world\n", string(data))
}
