package log

import (
	"errors"
	"testing"
)

type logrLogger interface {
	Enabled() bool
	Info(msg string, keysAndValues ...interface{})
	Error(err error, msg string, keysAndValues ...interface{})
	V(level int) *LogrLogger
	WithValues(keysAndValues ...interface{}) *LogrLogger
	WithName(name string) *LogrLogger
}

func TestLogrLoggerNil(t *testing.T) {
	var logger *Logger

	var logr logrLogger = logger.Logr(NewContext().Str("tag", "hi logr").Value())

	logr.Info("hello", "foo", "bar", "number", 42)
	logr.Error(errors.New("this is a error"), "hello", "foo", "bar", "number", 42)
	logr.V(0).Info("logr.V(0) OK.")

	logr = logr.WithName("a_named_logger")
	logr.Info("this is WithName test")

	logr = logr.WithValues("a_key", "a_value")
	logr.Info("this is WithValues test")

	if logr.Enabled() {
		logr.Info("logr.Enabled() OK.")
	}
}

func TestLogrLogger(t *testing.T) {
	DefaultLogger.Caller = 1
	DefaultLogger.Writer = &ConsoleWriter{ColorOutput: true, EndWithMessage: true}

	var logr logrLogger = DefaultLogger.Logr(NewContext().Str("tag", "hi logr").Value())

	logr.Info("hello", "foo", "bar", "number", 42)
	logr.Error(errors.New("this is a error"), "hello", "foo", "bar", "number", 42)
	logr.V(0).Info("logr.V(0) OK.")

	logr = logr.WithName("a_named_logger")
	logr.Info("this is WithName test")

	logr = logr.WithValues("a_key", "a_value")
	logr.Info("this is WithValues test")

	if logr.Enabled() {
		logr.Info("logr.Enabled() OK.")
	}
}

func TestLogrLoggerLevel(t *testing.T) {
	DefaultLogger.Caller = 1
	DefaultLogger.Writer = &ConsoleWriter{ColorOutput: true, EndWithMessage: true}

	var logr logrLogger = DefaultLogger.Logr(NewContext().Str("tag", "hi logr").Value())

	logr.(*LogrLogger).logger.Level = FatalLevel

	logr.Info("hello", "foo", "bar", "number", 42)
	logr.Error(errors.New("this is a error"), "hello", "foo", "bar", "number", 42)
}
