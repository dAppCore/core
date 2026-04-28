// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"os"
	"testing"

	. "dappco.re/go/core"
)

func TestOs_FileMode_Good_Alias(t *testing.T) {
	// FileMode alias must be assignable to/from os.FileMode.
	var coreMode FileMode = 0o644
	var osMode os.FileMode = coreMode
	AssertEqual(t, os.FileMode(0o644), osMode)
}

func TestOs_ModePerm_Good(t *testing.T) {
	AssertEqual(t, FileMode(0o777), ModePerm)
}

func TestOs_ModeDir_Good(t *testing.T) {
	mode := ModeDir | 0o755
	AssertTrue(t, mode.IsDir())
	AssertEqual(t, FileMode(0o755), mode.Perm())
}

func TestOs_Stdin_Good_NotNil(t *testing.T) {
	AssertNotNil(t, Stdin())
}

func TestOs_Stdout_Good_NotNil(t *testing.T) {
	AssertNotNil(t, Stdout())
}

func TestOs_Stderr_Good_NotNil(t *testing.T) {
	AssertNotNil(t, Stderr())
}
