// SPDX-License-Identifier: EUPL-1.2

package core

func TestApp_isExecutable_Good(t *T) {
	path := Path(t.TempDir(), "agent")
	RequireTrue(t, WriteFile(path, []byte("#!/bin/sh\n"), 0o755).OK)

	AssertTrue(t, isExecutable(path))
}
func TestApp_isExecutable_Bad(t *T) {
	AssertFalse(t, isExecutable(Path(t.TempDir(), "missing")))
}
func TestApp_isExecutable_Ugly(t *T) {
	AssertFalse(t, isExecutable(t.TempDir()))
}
