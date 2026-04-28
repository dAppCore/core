// SPDX-License-Identifier: EUPL-1.2

// Base OS primitives for the Core framework.
//
// Re-exports stdlib os value types and the standard streams so consumer
// packages declare FileMode parameters and reach Stdin/Stdout/Stderr
// without importing os directly.
//
// File operations live on c.Fs() (sandbox-aware). Environment lookups
// live on core.Env. Process control lives on c.Process() / go-process.
// What's here is the connecting tissue: types, constants, and the
// canonical stdio streams that boundary code can't avoid.
//
// Usage:
//
//	func writeMode(p string, mode core.FileMode) { ... }
//
//	core.WriteString(core.Stderr(), "diagnostic\n")
//
//	if mode.Perm()&core.ModePerm == 0o600 { ... }
package core

import "os"

// FileMode is an alias for os.FileMode — file mode bits and permissions.
//
//	mode := core.FileMode(0o600)
//	core.Println(mode.Perm())
type FileMode = os.FileMode

// File mode bits exposed at core scope. These are the same values as
// os.ModeDir etc., re-exported so consumers don't need to import os.
//
//	mode := core.ModeDir | core.ModePerm
//	if mode&core.ModeDir != 0 { core.Println("directory") }
const (
	ModeDir        = os.ModeDir
	ModeAppend     = os.ModeAppend
	ModeExclusive  = os.ModeExclusive
	ModeTemporary  = os.ModeTemporary
	ModeSymlink    = os.ModeSymlink
	ModeDevice     = os.ModeDevice
	ModeNamedPipe  = os.ModeNamedPipe
	ModeSocket     = os.ModeSocket
	ModeSetuid     = os.ModeSetuid
	ModeSetgid     = os.ModeSetgid
	ModeCharDevice = os.ModeCharDevice
	ModeSticky     = os.ModeSticky
	ModeIrregular  = os.ModeIrregular
	ModeType       = os.ModeType
	ModePerm       = os.ModePerm // 0o777 — Unix permission bits
)

// (Note: core.Signal is the existing Core primitive in signal.go for
// signal-event handling — distinct from os.Signal the interface. Use
// c.Signal() for the action-based signal surface.)

// Stdin returns the canonical standard input stream as an io.Reader.
//
//	scanner := core.NewLineScanner(core.Stdin())
func Stdin() Reader {
	return os.Stdin
}

// Stdout returns the canonical standard output stream as an io.Writer.
//
//	core.WriteString(core.Stdout(), "ready\n")
func Stdout() Writer {
	return os.Stdout
}

// Stderr returns the canonical standard error stream as an io.Writer.
//
//	core.WriteString(core.Stderr(), "warning\n")
func Stderr() Writer {
	return os.Stderr
}
