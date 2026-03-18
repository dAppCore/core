// SPDX-License-Identifier: EUPL-1.2

// Structured errors, crash recovery, and reporting for the Core framework.
// Provides E() for error creation, Wrap()/WrapCode() for chaining,
// and Err for panic recovery and crash reporting.

package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// ErrSink is the shared interface for error reporting.
// Implemented by ErrLog (structured logging) and ErrPan (panic recovery).
type ErrSink interface {
	Error(msg string, keyvals ...any)
	Warn(msg string, keyvals ...any)
}

var _ ErrSink = (*Log)(nil)

// Err represents a structured error with operational context.
// It implements the error interface and supports unwrapping.
type Err struct {
	Op   string // Operation being performed (e.g., "user.Save")
	Msg  string // Human-readable message
	Err  error  // Underlying error (optional)
	Code string // Error code (optional, e.g., "VALIDATION_FAILED")
}

// Error implements the error interface.
func (e *Err) Error() string {
	var prefix string
	if e.Op != "" {
		prefix = e.Op + ": "
	}
	if e.Err != nil {
		if e.Code != "" {
			return fmt.Sprintf("%s%s [%s]: %v", prefix, e.Msg, e.Code, e.Err)
		}
		return fmt.Sprintf("%s%s: %v", prefix, e.Msg, e.Err)
	}
	if e.Code != "" {
		return fmt.Sprintf("%s%s [%s]", prefix, e.Msg, e.Code)
	}
	return fmt.Sprintf("%s%s", prefix, e.Msg)
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *Err) Unwrap() error {
	return e.Err
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
	return &Err{Op: op, Msg: msg, Err: err}
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
		return &Err{Op: op, Msg: msg, Err: err, Code: logErr.Code}
	}
	return &Err{Op: op, Msg: msg, Err: err}
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
	return &Err{Op: op, Msg: msg, Err: err, Code: code}
}

// NewCode creates an error with just code and message (no underlying error).
// Useful for creating sentinel errors with codes.
//
// Example:
//
//	var ErrNotFound = log.NewCode("NOT_FOUND", "resource not found")
func NewCode(code, msg string) error {
	return &Err{Msg: msg, Code: code}
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

// Join combines multiple errors into one.
// Wrapper around errors.Join for convenience.
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// --- Error Introspection Helpers ---

// Op extracts the operation name from an error.
// Returns empty string if the error is not an *Err.
func Op(err error) string {
	var e *Err
	if As(err, &e) {
		return e.Op
	}
	return ""
}

// ErrCode extracts the error code from an error.
// Returns empty string if the error is not an *Err or has no code.
func ErrCode(err error) string {
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
		return e.Msg
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

// AllOps returns an iterator over all operational contexts in the error chain.
// It traverses the error tree using errors.Unwrap.
func AllOps(err error) iter.Seq[string] {
	return func(yield func(string) bool) {
		for err != nil {
			if e, ok := err.(*Err); ok {
				if e.Op != "" {
					if !yield(e.Op) {
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
	for op := range AllOps(err) {
		stack = append(stack, op)
	}
	return stack
}

// FormatStackTrace returns a pretty-printed logical stack trace.
func FormatStackTrace(err error) string {
	var ops []string
	for op := range AllOps(err) {
		ops = append(ops, op)
	}
	if len(ops) == 0 {
		return ""
	}
	return strings.Join(ops, " -> ")
}

// --- ErrLog: Log-and-Return Error Helpers ---

// ErrOpts holds shared options for error subsystems.
type ErrOpts struct {
	Log *Log
}

// ErrLog combines error creation with logging.
// Primary action: return an error. Secondary: log it.
type ErrLog struct {
	*ErrOpts
}

// NewErrLog creates an ErrLog (consumer convenience).
func NewErrLog(opts *ErrOpts) *ErrLog {
	return &ErrLog{opts}
}

func (el *ErrLog) log() *Log {
	if el.ErrOpts != nil && el.Log != nil {
		return el.Log
	}
	return defaultLog
}

// Error logs at Error level and returns a wrapped error.
func (el *ErrLog) Error(err error, op, msg string) error {
	if err == nil {
		return nil
	}
	wrapped := Wrap(err, op, msg)
	el.log().Error(msg, "op", op, "err", err)
	return wrapped
}

// Warn logs at Warn level and returns a wrapped error.
func (el *ErrLog) Warn(err error, op, msg string) error {
	if err == nil {
		return nil
	}
	wrapped := Wrap(err, op, msg)
	el.log().Warn(msg, "op", op, "err", err)
	return wrapped
}

// Must logs and panics if err is not nil.
func (el *ErrLog) Must(err error, op, msg string) {
	if err != nil {
		el.log().Error(msg, "op", op, "err", err)
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
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	Version string `json:"go_version"`
}

// ErrPan manages panic recovery and crash reporting.
type ErrPan struct {
	filePath string
	meta     map[string]string
	onCrash  func(CrashReport)
}

// PanOpts configures an ErrPan.
type PanOpts struct {
	// FilePath is the crash report JSON output path. Empty disables file output.
	FilePath string
	// Meta is metadata included in every crash report.
	Meta map[string]string
	// OnCrash is a callback invoked on every crash.
	OnCrash func(CrashReport)
}

// NewErrPan creates an ErrPan (consumer convenience).
func NewErrPan(opts ...PanOpts) *ErrPan {
	h := &ErrPan{}
	if len(opts) > 0 {
		o := opts[0]
		h.filePath = o.FilePath
		if o.Meta != nil {
			h.meta = maps.Clone(o.Meta)
		}
		h.onCrash = o.OnCrash
	}
	return h
}

// Recover captures a panic and creates a crash report.
// Use as: defer c.Error().Recover()
func (h *ErrPan) Recover() {
	if h == nil {
		return
	}
	r := recover()
	if r == nil {
		return
	}

	err, ok := r.(error)
	if !ok {
		err = fmt.Errorf("%v", r)
	}

	report := CrashReport{
		Timestamp: time.Now(),
		Error:     err.Error(),
		Stack:     string(debug.Stack()),
		System: CrashSystem{
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
			Version: runtime.Version(),
		},
		Meta: h.meta,
	}

	if h.onCrash != nil {
		h.onCrash(report)
	}

	if h.filePath != "" {
		h.appendReport(report)
	}
}

// SafeGo runs a function in a goroutine with panic recovery.
func (h *ErrPan) SafeGo(fn func()) {
	go func() {
		defer h.Recover()
		fn()
	}()
}

// Reports returns the last n crash reports from the file.
func (h *ErrPan) Reports(n int) ([]CrashReport, error) {
	if h.filePath == "" {
		return nil, nil
	}
	data, err := os.ReadFile(h.filePath)
	if err != nil {
		return nil, err
	}
	var reports []CrashReport
	if err := json.Unmarshal(data, &reports); err != nil {
		return nil, err
	}
	if n <= 0 || len(reports) <= n {
		return reports, nil
	}
	return reports[len(reports)-n:], nil
}

var crashMu sync.Mutex

func (h *ErrPan) appendReport(report CrashReport) {
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
