package log

import (
	"runtime"
)

// LogrLogger implements methods to satisfy interface
// github.com/go-logr/logr.Logger.
type LogrLogger struct {
	logger  Logger
	context Context
}

// Logr wraps the Logger to provide a logr logger
func (l *Logger) Logr(context Context) *LogrLogger {
	if l == nil {
		return nil
	}
	return &LogrLogger{
		logger:  *l,
		context: context,
	}
}

// Info logs a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func (l *LogrLogger) Info(msg string, keysAndValues ...interface{}) {
	if l == nil || l.logger.silent(InfoLevel) {
		return
	}
	e := l.logger.header(InfoLevel)
	if l.logger.Caller > 0 {
		_, file, line, _ := runtime.Caller(l.logger.Caller)
		e.caller(file, line, l.logger.Fullpath)
	}
	e.Context(l.context).KeysAndValues(keysAndValues...).Msg(msg)
}

// Error logs an error, with the given message and key/value pairs as context.
// It functions similarly to calling Info with the "error" named value, but may
// have unique behavior, and should be preferred for logging errors (see the
// package documentations for more information).
//
// The msg field should be used to add context to any underlying error,
// while the err field should be used to attach the actual error that
// triggered this log line, if present.
func (l *LogrLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	if l == nil || l.logger.silent(ErrorLevel) {
		return
	}
	e := l.logger.header(ErrorLevel)
	if l.logger.Caller > 0 {
		_, file, line, _ := runtime.Caller(l.logger.Caller)
		e.caller(file, line, l.logger.Fullpath)
	}
	e.Context(l.context).KeysAndValues(keysAndValues...).Msg(msg)
}

// Enabled tests whether this Logger is enabled.  For example, commandline
// flags might be used to set the logging verbosity and disable some info
// logs.
func (l *LogrLogger) Enabled() bool {
	return l != nil
}

// WithValues adds some key-value pairs of context to a logger.
// See Info for documentation on how key/value pairs work.
func (l *LogrLogger) WithValues(keysAndValues ...interface{}) *LogrLogger {
	if l == nil {
		return nil
	}
	l.context = append(l.context, NewContext(nil).KeysAndValues(keysAndValues...).Value()...)
	return l
}

// WithName adds a new element to the logger's name.
// Successive calls with WithName continue to append
// suffixes to the logger's name.  It's strongly recommended
// that name segments contain only letters, digits, and hyphens
// (see the package documentation for more information).
func (l *LogrLogger) WithName(name string) *LogrLogger {
	if l == nil {
		return nil
	}
	return l.WithValues("logger", name)
}

// V returns an Logger value for a specific verbosity level, relative to
// this Logger.  In other words, V values are additive.  V higher verbosity
// level means a log message is less important.  It's illegal to pass a log
// level less than zero.
func (l *LogrLogger) V(level int) *LogrLogger {
	if l != nil && int(l.logger.Level) <= level {
		return l
	}
	return nil
}
