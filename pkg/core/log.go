// Structured logging for the Core framework.
//
//	core.SetLevel(core.LevelDebug)
//	core.Info("server started", "port", 8080)
//	core.Error("failed to connect", "err", err)
package core

import (
	goio "io"
	"os"
	"os/user"
	"slices"
	"sync"
	"time"
)

// Level defines logging verbosity.
type Level int

// Logging level constants ordered by increasing verbosity.
const (
	// LevelQuiet suppresses all log output.
	LevelQuiet Level = iota
	// LevelError shows only error messages.
	LevelError
	// LevelWarn shows warnings and errors.
	LevelWarn
	// LevelInfo shows informational messages, warnings, and errors.
	LevelInfo
	// LevelDebug shows all messages including debug details.
	LevelDebug
)

// String returns the level name.
func (l Level) String() string {
	switch l {
	case LevelQuiet:
		return "quiet"
	case LevelError:
		return "error"
	case LevelWarn:
		return "warn"
	case LevelInfo:
		return "info"
	case LevelDebug:
		return "debug"
	default:
		return "unknown"
	}
}

// Log provides structured logging.
type Log struct {
	mu     sync.RWMutex
	level  Level
	output goio.Writer

	// RedactKeys is a list of keys whose values should be masked in logs.
	redactKeys []string

	// Style functions for formatting (can be overridden)
	StyleTimestamp func(string) string
	StyleDebug     func(string) string
	StyleInfo      func(string) string
	StyleWarn      func(string) string
	StyleError     func(string) string
	StyleSecurity  func(string) string
}

// RotationLogOptions defines the log rotation and retention policy.
type RotationLogOptions struct {
	// Filename is the log file path. If empty, rotation is disabled.
	Filename string

	// MaxSize is the maximum size of the log file in megabytes before it gets rotated.
	// It defaults to 100 megabytes.
	MaxSize int

	// MaxAge is the maximum number of days to retain old log files based on their
	// file modification time. It defaults to 28 days.
	// Note: set to a negative value to disable age-based retention.
	MaxAge int

	// MaxBackups is the maximum number of old log files to retain.
	// It defaults to 5 backups.
	MaxBackups int

	// Compress determines if the rotated log files should be compressed using gzip.
	// It defaults to true.
	Compress bool
}

// LogOptions configures a Log.
type LogOptions struct {
	Level Level
	// Output is the destination for log messages. If Rotation is provided,
	// Output is ignored and logs are written to the rotating file instead.
	Output goio.Writer
	// Rotation enables log rotation to file. If provided, Filename must be set.
	Rotation *RotationLogOptions
	// RedactKeys is a list of keys whose values should be masked in logs.
	RedactKeys []string
}

// RotationWriterFactory creates a rotating writer from options.
// Set this to enable log rotation (provided by core/go-io integration).
var RotationWriterFactory func(RotationLogOptions) goio.WriteCloser

// New creates a new Log with the given options.
func NewLog(opts LogOptions) *Log {
	output := opts.Output
	if opts.Rotation != nil && opts.Rotation.Filename != "" && RotationWriterFactory != nil {
		output = RotationWriterFactory(*opts.Rotation)
	}
	if output == nil {
		output = os.Stderr
	}

	return &Log{
		level:          opts.Level,
		output:         output,
		redactKeys:     slices.Clone(opts.RedactKeys),
		StyleTimestamp: identity,
		StyleDebug:     identity,
		StyleInfo:      identity,
		StyleWarn:      identity,
		StyleError:     identity,
		StyleSecurity:  identity,
	}
}

func identity(s string) string { return s }

// SetLevel changes the log level.
func (l *Log) SetLevel(level Level) {
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}

// Level returns the current log level.
func (l *Log) Level() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// SetOutput changes the output writer.
func (l *Log) SetOutput(w goio.Writer) {
	l.mu.Lock()
	l.output = w
	l.mu.Unlock()
}

// SetRedactKeys sets the keys to be redacted.
func (l *Log) SetRedactKeys(keys ...string) {
	l.mu.Lock()
	l.redactKeys = slices.Clone(keys)
	l.mu.Unlock()
}

func (l *Log) shouldLog(level Level) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return level <= l.level
}

func (l *Log) log(level Level, prefix, msg string, keyvals ...any) {
	l.mu.RLock()
	output := l.output
	styleTimestamp := l.StyleTimestamp
	redactKeys := l.redactKeys
	l.mu.RUnlock()

	timestamp := styleTimestamp(time.Now().Format("15:04:05"))

	// Copy keyvals to avoid mutating the caller's slice
	keyvals = append([]any(nil), keyvals...)

	// Automatically extract context from error if present in keyvals
	origLen := len(keyvals)
	for i := 0; i < origLen; i += 2 {
		if i+1 < origLen {
			if err, ok := keyvals[i+1].(error); ok {
				if op := Operation(err); op != "" {
					// Check if op is already in keyvals
					hasOp := false
					for j := 0; j < len(keyvals); j += 2 {
						if k, ok := keyvals[j].(string); ok && k == "op" {
							hasOp = true
							break
						}
					}
					if !hasOp {
						keyvals = append(keyvals, "op", op)
					}
				}
				if stack := FormatStackTrace(err); stack != "" {
					// Check if stack is already in keyvals
					hasStack := false
					for j := 0; j < len(keyvals); j += 2 {
						if k, ok := keyvals[j].(string); ok && k == "stack" {
							hasStack = true
							break
						}
					}
					if !hasStack {
						keyvals = append(keyvals, "stack", stack)
					}
				}
			}
		}
	}

	// Format key-value pairs
	var kvStr string
	if len(keyvals) > 0 {
		kvStr = " "
		for i := 0; i < len(keyvals); i += 2 {
			if i > 0 {
				kvStr += " "
			}
			key := keyvals[i]
			var val any
			if i+1 < len(keyvals) {
				val = keyvals[i+1]
			}

			// Redaction logic
			keyStr := Sprint(key)
			if slices.Contains(redactKeys, keyStr) {
				val = "[REDACTED]"
			}

			// Secure formatting to prevent log injection
			if s, ok := val.(string); ok {
				kvStr += Sprintf("%v=%q", key, s)
			} else {
				kvStr += Sprintf("%v=%v", key, val)
			}
		}
	}

	Print(output, "%s %s %s%s", timestamp, prefix, msg, kvStr)
}

// Debug logs a debug message with optional key-value pairs.
func (l *Log) Debug(msg string, keyvals ...any) {
	if l.shouldLog(LevelDebug) {
		l.log(LevelDebug, l.StyleDebug("[DBG]"), msg, keyvals...)
	}
}

// Info logs an info message with optional key-value pairs.
func (l *Log) Info(msg string, keyvals ...any) {
	if l.shouldLog(LevelInfo) {
		l.log(LevelInfo, l.StyleInfo("[INF]"), msg, keyvals...)
	}
}

// Warn logs a warning message with optional key-value pairs.
func (l *Log) Warn(msg string, keyvals ...any) {
	if l.shouldLog(LevelWarn) {
		l.log(LevelWarn, l.StyleWarn("[WRN]"), msg, keyvals...)
	}
}

// Error logs an error message with optional key-value pairs.
func (l *Log) Error(msg string, keyvals ...any) {
	if l.shouldLog(LevelError) {
		l.log(LevelError, l.StyleError("[ERR]"), msg, keyvals...)
	}
}

// Security logs a security event with optional key-value pairs.
// It uses LevelError to ensure security events are visible even in restrictive
// log configurations.
func (l *Log) Security(msg string, keyvals ...any) {
	if l.shouldLog(LevelError) {
		l.log(LevelError, l.StyleSecurity("[SEC]"), msg, keyvals...)
	}
}

// Username returns the current system username.
// It uses os/user for reliability and falls back to environment variables.
func Username() string {
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	// Fallback for environments where user lookup might fail
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	return os.Getenv("USERNAME")
}

// --- Default logger ---

var defaultLog = NewLog(LogOptions{Level: LevelInfo})

// Default returns the default logger.
func Default() *Log {
	return defaultLog
}

// SetDefault sets the default logger.
func SetDefault(l *Log) {
	defaultLog = l
}

// SetLevel sets the default logger's level.
func SetLevel(level Level) {
	defaultLog.SetLevel(level)
}

// SetRedactKeys sets the default logger's redaction keys.
func SetRedactKeys(keys ...string) {
	defaultLog.SetRedactKeys(keys...)
}

// Debug logs to the default logger.
func Debug(msg string, keyvals ...any) {
	defaultLog.Debug(msg, keyvals...)
}

// Info logs to the default logger.
func Info(msg string, keyvals ...any) {
	defaultLog.Info(msg, keyvals...)
}

// Warn logs to the default logger.
func Warn(msg string, keyvals ...any) {
	defaultLog.Warn(msg, keyvals...)
}

// Error logs to the default logger.
func Error(msg string, keyvals ...any) {
	defaultLog.Error(msg, keyvals...)
}

// Security logs to the default logger.
func Security(msg string, keyvals ...any) {
	defaultLog.Security(msg, keyvals...)
}

// --- LogErr: Error-Aware Logger ---

// LogErr logs structured information extracted from errors.
// Primary action: log. Secondary: extract error context.
type LogErr struct {
	log *Log
}

// NewLogErr creates a LogErr bound to the given logger.
func NewLogErr(log *Log) *LogErr {
	return &LogErr{log: log}
}

// Log extracts context from an Err and logs it at Error level.
func (le *LogErr) Log(err error) {
	if err == nil {
		return
	}
	le.log.Error(ErrorMessage(err), "op", Operation(err), "code", ErrorCode(err), "stack", FormatStackTrace(err))
}

// --- LogPanic: Panic-Aware Logger ---

// LogPanic logs panic context without crash file management.
// Primary action: log. Secondary: recover panics.
type LogPanic struct {
	log *Log
}

// NewLogPanic creates a LogPanic bound to the given logger.
func NewLogPanic(log *Log) *LogPanic {
	return &LogPanic{log: log}
}

// Recover captures a panic and logs it. Does not write crash files.
// Use as: defer core.NewLogPanic(logger).Recover()
func (lp *LogPanic) Recover() {
	r := recover()
	if r == nil {
		return
	}
	err, ok := r.(error)
	if !ok {
		err = NewError(Sprint("panic: ", r))
	}
	lp.log.Error("panic recovered",
		"err", err,
		"op", Operation(err),
		"stack", FormatStackTrace(err),
	)
}
