// Structured logging for the Core framework.
//
//	core.SetLevel(core.LevelDebug)
//	core.Info("server started", "port", 8080)
//	core.Error("failed to connect", "err", err)
package core

import (
	goio "io"
	"os"
	"sync"
)

// Level defines logging verbosity.
//
//	level := core.LevelInfo
//	core.SetLevel(level)
type Level int

// Logging level constants ordered by increasing verbosity.
const (
	// LevelQuiet suppresses all log output.
	//
	//	core.SetLevel(core.LevelQuiet)
	LevelQuiet Level = iota
	// LevelError shows only error messages.
	//
	//	core.SetLevel(core.LevelError)
	//	core.Error("agent failed", "service", "homelab")
	LevelError
	// LevelWarn shows warnings and errors.
	//
	//	core.SetLevel(core.LevelWarn)
	//	core.Warn("agent degraded", "service", "homelab")
	LevelWarn
	// LevelInfo shows informational messages, warnings, and errors.
	//
	//	core.SetLevel(core.LevelInfo)
	//	core.Info("agent started", "service", "homelab")
	LevelInfo
	// LevelDebug shows all messages including debug details.
	//
	//	core.SetLevel(core.LevelDebug)
	//	core.Debug("agent trace", "task", "task-42")
	LevelDebug
)

// String returns the level name.
//
//	name := core.LevelInfo.String()
//	core.Println(name)
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
//
//	log := core.NewLog(core.LogOptions{Level: core.LevelInfo, Output: core.Stdout()})
//	log.Info("agent started", "service", "homelab")
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
//
//	opts := core.RotationLogOptions{Filename: "/var/log/core/agent.log", MaxSize: 100, MaxBackups: 5, Compress: true}
//	_ = opts
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
//
//	opts := core.LogOptions{Level: core.LevelInfo, Output: core.Stdout(), RedactKeys: []string{"token"}}
//	log := core.NewLog(opts)
//	log.Info("agent started")
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
//
//	core.RotationWriterFactory = func(opts core.RotationLogOptions) core.WriteCloser {
//	    return nil
//	}
var RotationWriterFactory func(RotationLogOptions) goio.WriteCloser

// New creates a new Log with the given options.
//
//	log := core.NewLog(core.LogOptions{Level: core.LevelDebug, Output: core.Stdout()})
//	log.Debug("agent trace", "task", "task-42")
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
		redactKeys:     SliceClone(opts.RedactKeys),
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
//
//	log := core.NewLog(core.LogOptions{Output: core.Stdout()})
//	log.SetLevel(core.LevelDebug)
func (l *Log) SetLevel(level Level) {
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}

// Level returns the current log level.
//
//	log := core.NewLog(core.LogOptions{Level: core.LevelInfo})
//	level := log.Level()
//	core.Println(level.String())
func (l *Log) Level() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// SetOutput changes the output writer.
//
//	log := core.NewLog(core.LogOptions{Level: core.LevelInfo})
//	log.SetOutput(core.Stdout())
func (l *Log) SetOutput(w goio.Writer) {
	l.mu.Lock()
	l.output = w
	l.mu.Unlock()
}

// SetRedactKeys sets the keys to be redacted.
//
//	log := core.NewLog(core.LogOptions{Level: core.LevelInfo})
//	log.SetRedactKeys("token", "authorization")
func (l *Log) SetRedactKeys(keys ...string) {
	l.mu.Lock()
	l.redactKeys = SliceClone(keys)
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

	timestamp := styleTimestamp(Now().Format("15:04:05"))

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
			if SliceContains(redactKeys, keyStr) {
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
//
//	log := core.NewLog(core.LogOptions{Level: core.LevelDebug, Output: core.Stdout()})
//	log.Debug("agent trace", "task", "task-42")
func (l *Log) Debug(msg string, keyvals ...any) {
	if l.shouldLog(LevelDebug) {
		l.log(LevelDebug, l.StyleDebug("[DBG]"), msg, keyvals...)
	}
}

// Info logs an info message with optional key-value pairs.
//
//	log := core.NewLog(core.LogOptions{Level: core.LevelInfo, Output: core.Stdout()})
//	log.Info("agent started", "service", "homelab")
func (l *Log) Info(msg string, keyvals ...any) {
	if l.shouldLog(LevelInfo) {
		l.log(LevelInfo, l.StyleInfo("[INF]"), msg, keyvals...)
	}
}

// Warn logs a warning message with optional key-value pairs.
//
//	log := core.NewLog(core.LogOptions{Level: core.LevelWarn, Output: core.Stdout()})
//	log.Warn("agent degraded", "service", "homelab")
func (l *Log) Warn(msg string, keyvals ...any) {
	if l.shouldLog(LevelWarn) {
		l.log(LevelWarn, l.StyleWarn("[WRN]"), msg, keyvals...)
	}
}

// Error logs an error message with optional key-value pairs.
//
//	log := core.NewLog(core.LogOptions{Level: core.LevelError, Output: core.Stdout()})
//	log.Error("agent failed", "service", "homelab")
func (l *Log) Error(msg string, keyvals ...any) {
	if l.shouldLog(LevelError) {
		l.log(LevelError, l.StyleError("[ERR]"), msg, keyvals...)
	}
}

// Security logs a security event with optional key-value pairs.
// It uses LevelError to ensure security events are visible even in restrictive
// log configurations.
//
//	log := core.NewLog(core.LogOptions{Level: core.LevelError, Output: core.Stdout()})
//	log.Security("entitlement.denied", "action", "process.run")
func (l *Log) Security(msg string, keyvals ...any) {
	if l.shouldLog(LevelError) {
		l.log(LevelError, l.StyleSecurity("[SEC]"), msg, keyvals...)
	}
}

// Username returns the current system username, falling back to USER /
// USERNAME env vars when user lookup fails (e.g. distroless containers
// without a /etc/passwd entry).
//
//	user := core.Username()
//	core.Println(user)
func Username() string {
	if r := UserCurrent(); r.OK {
		return r.Value.(*User).Username
	}
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	return os.Getenv("USERNAME")
}

// --- Default logger ---

var defaultLogPtr AtomicPointer[Log]

func init() {
	l := NewLog(LogOptions{Level: LevelInfo})
	defaultLogPtr.Store(l)
}

// Default returns the default logger.
//
//	log := core.Default()
//	log.Info("agent started")
func Default() *Log {
	return defaultLogPtr.Load()
}

// SetDefault sets the default logger.
//
//	log := core.NewLog(core.LogOptions{Level: core.LevelInfo, Output: core.Stdout()})
//	core.SetDefault(log)
func SetDefault(l *Log) {
	defaultLogPtr.Store(l)
}

// SetLevel sets the default logger's level.
//
//	core.SetLevel(core.LevelWarn)
func SetLevel(level Level) {
	Default().SetLevel(level)
}

// SetRedactKeys sets the default logger's redaction keys.
//
//	core.SetRedactKeys("token", "authorization")
func SetRedactKeys(keys ...string) {
	Default().SetRedactKeys(keys...)
}

// Debug logs to the default logger.
//
//	core.SetLevel(core.LevelDebug)
//	core.Debug("agent trace", "task", "task-42")
func Debug(msg string, keyvals ...any) {
	Default().Debug(msg, keyvals...)
}

// Info logs to the default logger.
//
//	core.Info("agent started", "service", "homelab")
func Info(msg string, keyvals ...any) {
	Default().Info(msg, keyvals...)
}

// Warn logs to the default logger.
//
//	core.Warn("agent degraded", "service", "homelab")
func Warn(msg string, keyvals ...any) {
	Default().Warn(msg, keyvals...)
}

// Error logs to the default logger.
//
//	core.Error("agent failed", "service", "homelab")
func Error(msg string, keyvals ...any) {
	Default().Error(msg, keyvals...)
}

// Security logs to the default logger.
//
//	core.Security("entitlement.denied", "action", "process.run")
func Security(msg string, keyvals ...any) {
	Default().Security(msg, keyvals...)
}

// --- LogErr: Error-Aware Logger ---

// LogErr logs structured information extracted from errors.
// Primary action: log. Secondary: extract error context.
//
//	logger := core.NewLogErr(core.Default())
//	logger.Log(core.E("agent.Run", "failed", nil))
type LogErr struct {
	log *Log
}

// NewLogErr creates a LogErr bound to the given logger.
//
//	logger := core.NewLogErr(core.Default())
//	logger.Log(core.E("agent.Run", "failed", nil))
func NewLogErr(log *Log) *LogErr {
	return &LogErr{log: log}
}

// Log extracts context from an Err and logs it at Error level.
//
//	logger := core.NewLogErr(core.Default())
//	logger.Log(core.E("agent.Run", "failed", nil))
func (le *LogErr) Log(err error) {
	if err == nil {
		return
	}
	le.log.Error(ErrorMessage(err), "op", Operation(err), "code", ErrorCode(err), "stack", FormatStackTrace(err))
}

// --- LogPanic: Panic-Aware Logger ---

// LogPanic logs panic context without crash file management.
// Primary action: log. Secondary: recover panics.
//
//	guard := core.NewLogPanic(core.Default())
//	defer guard.Recover()
type LogPanic struct {
	log *Log
}

// NewLogPanic creates a LogPanic bound to the given logger.
//
//	guard := core.NewLogPanic(core.Default())
//	defer guard.Recover()
func NewLogPanic(log *Log) *LogPanic {
	return &LogPanic{log: log}
}

// Recover captures a panic and logs it. Does not write crash files.
// Use as: defer core.NewLogPanic(logger).Recover()
//
//	guard := core.NewLogPanic(core.Default())
//	defer guard.Recover()
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
