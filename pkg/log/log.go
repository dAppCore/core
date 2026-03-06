// Package log re-exports go-log and provides framework integration (Service)
// and log rotation (RotatingWriter) that depend on core/go internals.
//
// New code should import forge.lthn.ai/core/go-log directly.
package log

import (
	"io"

	golog "forge.lthn.ai/core/go-log"
)

// Type aliases — all go-log types available as log.X
type (
	Level           = golog.Level
	Logger          = golog.Logger
	Options         = golog.Options
	RotationOptions = golog.RotationOptions
	Err             = golog.Err
)

// Level constants.
const (
	LevelQuiet = golog.LevelQuiet
	LevelError = golog.LevelError
	LevelWarn  = golog.LevelWarn
	LevelInfo  = golog.LevelInfo
	LevelDebug = golog.LevelDebug
)

func init() {
	// Wire rotation into go-log: when go-log's New() gets RotationOptions,
	// it calls this factory to create the RotatingWriter (which needs go-io).
	golog.RotationWriterFactory = func(opts RotationOptions) io.WriteCloser {
		return NewRotatingWriter(opts, nil)
	}
}

// --- Logging functions (re-exported from go-log) ---

var (
	New        = golog.New
	Default    = golog.Default
	SetDefault = golog.SetDefault
	SetLevel   = golog.SetLevel
	Debug      = golog.Debug
	Info       = golog.Info
	Warn       = golog.Warn
	Error      = golog.Error
	Security   = golog.Security
	Username   = golog.Username
)

// --- Error functions (re-exported from go-log) ---

var (
	E               = golog.E
	Wrap            = golog.Wrap
	WrapCode        = golog.WrapCode
	NewCode         = golog.NewCode
	Is              = golog.Is
	As              = golog.As
	NewError        = golog.NewError
	Join            = golog.Join
	Op              = golog.Op
	ErrCode         = golog.ErrCode
	Message         = golog.Message
	Root            = golog.Root
	StackTrace      = golog.StackTrace
	FormatStackTrace = golog.FormatStackTrace
	LogError        = golog.LogError
	LogWarn         = golog.LogWarn
	Must            = golog.Must
)
