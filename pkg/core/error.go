// SPDX-License-Identifier: EUPL-1.2

// Structured errors, crash recovery, and reporting for the Core framework.
// Provides E() for error creation, Wrap()/WrapCode() for chaining,
// and Err for panic recovery and crash reporting.

package core

import (
	"encoding/json"
	"errors"
	"iter"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// ErrorSink is the shared interface for error reporting.
// Implemented by ErrorLog (structured logging) and ErrorPanic (panic recovery).
type ErrorSink interface {
	Error(msg string, keyvals ...any)
	Warn(msg string, keyvals ...any)
}

var _ ErrorSink = (*Log)(nil)

// Err represents a structured error with operational context.
// It implements the error interface and supports unwrapping.
type Err struct {
	Operation string // Operation being performed (e.g., "user.Save")
	Message   string // Human-readable message
	Cause     error  // Underlying error (optional)
	Code      string // Error code (optional, e.g., "VALIDATION_FAILED")
}

// Error implements the error interface.
func (e *Err) Error() string {
	var prefix string
	if e.Operation != "" {
		prefix = e.Operation + ": "
	}
	if e.Cause != nil {
		if e.Code != "" {
			return Concat(prefix, e.Message, " [", e.Code, "]: ", e.Cause.Error())
		}
		return Concat(prefix, e.Message, ": ", e.Cause.Error())
	}
	if e.Code != "" {
		return Concat(prefix, e.Message, " [", e.Code, "]")
	}
	return Concat(prefix, e.Message)
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *Err) Unwrap() error {
	return e.Cause
}

// --- Error Creation Functions ---

// E creates a new Err with operation context.
// The underlying error can be nil for creating errors without a cause.
//
// Example:
//
//	return log.E("user.Save", "failed to save user", err)
//	return log.E("api.Call", "rate limited", nil)  // No underlying cause
func E(op, msg string, err error) error {
	return &Err{Operation: op, Message: msg, Cause: err}
}

// Wrap wraps an error with operation context.
// Returns nil if err is nil, to support conditional wrapping.
// Preserves error Code if the wrapped error is an *Err.
//
// Example:
//
//	return log.Wrap(err, "db.Query", "database query failed")
func Wrap(err error, op, msg string) error {
	if err == nil {
		return nil
	}
	// Preserve Code from wrapped *Err
	var logErr *Err
	if As(err, &logErr) && logErr.Code != "" {
		return &Err{Operation: op, Message: msg, Cause: err, Code: logErr.Code}
	}
	return &Err{Operation: op, Message: msg, Cause: err}
}

// WrapCode wraps an error with operation context and error code.
// Returns nil only if both err is nil AND code is empty.
// Useful for API errors that need machine-readable codes.
//
// Example:
//
//	return log.WrapCode(err, "VALIDATION_ERROR", "user.Validate", "invalid email")
func WrapCode(err error, code, op, msg string) error {
	if err == nil && code == "" {
		return nil
	}
	return &Err{Operation: op, Message: msg, Cause: err, Code: code}
}

// NewCode creates an error with just code and message (no underlying error).
// Useful for creating sentinel errors with codes.
//
// Example:
//
//	var ErrNotFound = log.NewCode("NOT_FOUND", "resource not found")
func NewCode(code, msg string) error {
	return &Err{Message: msg, Code: code}
}

// --- Standard Library Wrappers ---

// Is reports whether any error in err's tree matches target.
// Wrapper around errors.Is for convenience.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's tree that matches target.
// Wrapper around errors.As for convenience.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// NewError creates a simple error with the given text.
// Wrapper around errors.New for convenience.
func NewError(text string) error {
	return errors.New(text)
}

// ErrorJoin combines multiple errors into one.
//
//	core.CauseorJoin(err1, err2, err3)
func ErrorJoin(errs ...error) error {
	return errors.Join(errs...)
}

// --- Error Introspection Helpers ---

// Operation extracts the operation name from an error.
// Returns empty string if the error is not an *Err.
func Operation(err error) string {
	var e *Err
	if As(err, &e) {
		return e.Operation
	}
	return ""
}

// ErrorCode extracts the error code from an error.
// Returns empty string if the error is not an *Err or has no code.
func ErrorCode(err error) string {
	var e *Err
	if As(err, &e) {
		return e.Code
	}
	return ""
}

// Message extracts the message from an error.
// Returns the error's Error() string if not an *Err.
func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	var e *Err
	if As(err, &e) {
		return e.Message
	}
	return err.Error()
}

// Root returns the root cause of an error chain.
// Unwraps until no more wrapped errors are found.
func Root(err error) error {
	if err == nil {
		return nil
	}
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

// AllOperations returns an iterator over all operational contexts in the error chain.
// It traverses the error tree using errors.Unwrap.
func AllOperations(err error) iter.Seq[string] {
	return func(yield func(string) bool) {
		for err != nil {
			if e, ok := err.(*Err); ok {
				if e.Operation != "" {
					if !yield(e.Operation) {
						return
					}
				}
			}
			err = errors.Unwrap(err)
		}
	}
}

// StackTrace returns the logical stack trace (chain of operations) from an error.
// It returns an empty slice if no operational context is found.
func StackTrace(err error) []string {
	var stack []string
	for op := range AllOperations(err) {
		stack = append(stack, op)
	}
	return stack
}

// FormatStackTrace returns a pretty-printed logical stack trace.
func FormatStackTrace(err error) string {
	var ops []string
	for op := range AllOperations(err) {
		ops = append(ops, op)
	}
	if len(ops) == 0 {
		return ""
	}
	return Join(" -> ", ops...)
}

// --- ErrorLog: Log-and-Return Error Helpers ---

// ErrorLog combines error creation with logging.
// Primary action: return an error. Secondary: log it.
type ErrorLog struct {
	log *Log
}

func (el *ErrorLog) logger() *Log {
	if el.log != nil {
		return el.log
	}
	return Default()
}

// Error logs at Error level and returns a Result with the wrapped error.
func (el *ErrorLog) Error(err error, op, msg string) Result {
	if err == nil {
		return Result{OK: true}
	}
	wrapped := Wrap(err, op, msg)
	el.logger().Error(msg, "op", op, "err", err)
	return Result{wrapped, false}
}

// Warn logs at Warn level and returns a Result with the wrapped error.
func (el *ErrorLog) Warn(err error, op, msg string) Result {
	if err == nil {
		return Result{OK: true}
	}
	wrapped := Wrap(err, op, msg)
	el.logger().Warn(msg, "op", op, "err", err)
	return Result{wrapped, false}
}

// Must logs and panics if err is not nil.
func (el *ErrorLog) Must(err error, op, msg string) {
	if err != nil {
		el.logger().Error(msg, "op", op, "err", err)
		panic(Wrap(err, op, msg))
	}
}

// --- Crash Recovery & Reporting ---

// CrashReport represents a single crash event.
type CrashReport struct {
	Timestamp time.Time         `json:"timestamp"`
	Error     string            `json:"error"`
	Stack     string            `json:"stack"`
	System    CrashSystem       `json:"system,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// CrashSystem holds system information at crash time.
type CrashSystem struct {
	OperatingSystem string `json:"operatingsystem"`
	Architecture    string `json:"architecture"`
	Version         string `json:"go_version"`
}

// ErrorPanic manages panic recovery and crash reporting.
type ErrorPanic struct {
	filePath string
	meta     map[string]string
	onCrash  func(CrashReport)
}

// Recover captures a panic and creates a crash report.
// Use as: defer c.Error().Recover()
func (h *ErrorPanic) Recover() {
	if h == nil {
		return
	}
	r := recover()
	if r == nil {
		return
	}

	err, ok := r.(error)
	if !ok {
		err = NewError(Sprint("panic: ", r))
	}

	report := CrashReport{
		Timestamp: time.Now(),
		Error:     err.Error(),
		Stack:     string(debug.Stack()),
		System: CrashSystem{
			OperatingSystem: runtime.GOOS,
			Architecture:    runtime.GOARCH,
			Version:         runtime.Version(),
		},
		Meta: maps.Clone(h.meta),
	}

	if h.onCrash != nil {
		h.onCrash(report)
	}

	if h.filePath != "" {
		h.appendReport(report)
	}
}

// SafeGo runs a function in a goroutine with panic recovery.
func (h *ErrorPanic) SafeGo(fn func()) {
	go func() {
		defer h.Recover()
		fn()
	}()
}

// Reports returns the last n crash reports from the file.
func (h *ErrorPanic) Reports(n int) Result {
	if h.filePath == "" {
		return Result{}
	}
	crashMu.Lock()
	defer crashMu.Unlock()
	data, err := os.ReadFile(h.filePath)
	if err != nil {
		return Result{err, false}
	}
	var reports []CrashReport
	if err := json.Unmarshal(data, &reports); err != nil {
		return Result{err, false}
	}
	if n <= 0 || len(reports) <= n {
		return Result{reports, true}
	}
	return Result{reports[len(reports)-n:], true}
}

var crashMu sync.Mutex

func (h *ErrorPanic) appendReport(report CrashReport) {
	crashMu.Lock()
	defer crashMu.Unlock()

	var reports []CrashReport
	if data, err := os.ReadFile(h.filePath); err == nil {
		if err := json.Unmarshal(data, &reports); err != nil {
			reports = nil
		}
	}

	reports = append(reports, report)
	if data, err := json.MarshalIndent(reports, "", "  "); err == nil {
		_ = os.MkdirAll(filepath.Dir(h.filePath), 0755)
		_ = os.WriteFile(h.filePath, data, 0600)
	}
}
