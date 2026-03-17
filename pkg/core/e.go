// Package core re-exports the structured error types from go-log.
//
// All error construction in the framework MUST use E() (or Wrap, WrapCode, etc.)
// rather than fmt.Errorf. This ensures every error carries an operation context
// for structured logging and tracing.
//
// Example:
//
//	return core.E("config.Load", "failed to load config file", err)
package core

import (
	coreerr "forge.lthn.ai/core/go-log"
)

// Error is the structured error type from go-log.
// It carries Op (operation), Msg (human-readable), Err (underlying), and Code fields.
type Error = coreerr.Err

// E creates a new structured error with operation context.
// This is the primary way to create errors in the Core framework.
//
// The 'op' parameter should be in the format of 'package.function' or 'service.method'.
// The 'msg' parameter should be a human-readable message.
// The 'err' parameter is the underlying error (may be nil).
var E = coreerr.E
