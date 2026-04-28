// SPDX-License-Identifier: EUPL-1.2

package core

import (
	"os"
	"os/exec"
)

// Helpers shared by *_internal_test.go files in this package.

var ErrPermissionForTest = os.ErrPermission

func SymlinkForTest(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

func ExecCmdForTest(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

type ax7FailingWriter struct{}

func (ax7FailingWriter) Write(_ []byte) (int, error) {
	return 0, E("ax7.failingWriter", "write failed", nil)
}

func ax7CrashReport(message string) CrashReport {
	return CrashReport{
		Timestamp: Now(),
		Error:     message,
		Stack:     "agent stack",
		Meta:      map[string]string{"agent": "codex"},
	}
}
