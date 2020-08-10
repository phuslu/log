package log

import (
	"fmt"
	"runtime"
	"unsafe"
)

// A SugaredLogger wraps the base Logger functionality in a slower, but less
// verbose, API. Any Logger can be converted to a SugaredLogger with its Sugar
// method.
//
// Unlike the Logger, the SugaredLogger doesn't insist on structured logging.
// For each log level, it exposes three methods: one for loosely-typed
// structured logging, one for println-style formatting, and one for
// printf-style formatting.
type SugaredLogger struct {
	logger  Logger
	context Context
}

// Sugar wraps the Logger to provide a more ergonomic, but a little bit slower,
// API. Sugaring a Logger is quite inexpensive, so it's reasonable for a
// single application to use both Loggers and SugaredLoggers, converting
// between them on the boundaries of performance-sensitive code.
func (l *Logger) Sugar(context Context) (s *SugaredLogger) {
	s = &SugaredLogger{
		logger:  *l,
		context: context,
	}
	return
}

// Level creates a child logger with the minimum accepted level set to level.
func (s *SugaredLogger) Level(level Level) *SugaredLogger {
	sl := *s
	sl.logger.Level = level
	return &sl
}

// Print sends a log event without extra field. Arguments are handled in the manner of fmt.Print.
func (s *SugaredLogger) Print(args ...interface{}) {
	e := s.logger.header(s.logger.Level)
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	print(e.Context(s.context), args)
}

// Println sends a log event without extra field. Arguments are handled in the manner of fmt.Print.
func (s *SugaredLogger) Println(args ...interface{}) {
	e := s.logger.header(s.logger.Level)
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	print(e.Context(s.context), args)
}

// Printf sends a log event without extra field. Arguments are handled in the manner of fmt.Printf.
func (s *SugaredLogger) Printf(format string, args ...interface{}) {
	e := s.logger.header(s.logger.Level)
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	e.Context(s.context).Msgf(format, args...)
}

// Log sends a log event with extra fields.
func (s *SugaredLogger) Log(keysAndValues ...interface{}) error {
	e := s.logger.header(s.logger.Level)
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	e.Context(s.context).keysAndValues(keysAndValues...).Msg("")
	return nil
}

// Debug uses fmt.Sprint to construct and log a message.
func (s *SugaredLogger) Debug(args ...interface{}) {
	e := s.logger.header(DebugLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	print(e.Context(s.context), args)
}

// Debugf uses fmt.Sprintf to log a templated message.
func (s *SugaredLogger) Debugf(template string, args ...interface{}) {
	e := s.logger.header(DebugLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	e.Context(s.context).Msgf(template, args...)
}

// Debugw logs a message with some additional context.
func (s *SugaredLogger) Debugw(msg string, keysAndValues ...interface{}) {
	e := s.logger.header(DebugLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	e.Context(s.context).keysAndValues(keysAndValues...).Msg(msg)
}

// Info uses fmt.Sprint to construct and log a message.
func (s *SugaredLogger) Info(args ...interface{}) {
	e := s.logger.header(InfoLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	print(e.Context(s.context), args)
}

// Infof uses fmt.Sprintf to log a templated message.
func (s *SugaredLogger) Infof(template string, args ...interface{}) {
	e := s.logger.header(InfoLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	e.Context(s.context).Msgf(template, args...)
}

// Infow logs a message with some additional context.
func (s *SugaredLogger) Infow(msg string, keysAndValues ...interface{}) {
	e := s.logger.header(InfoLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	e.Context(s.context).keysAndValues(keysAndValues...).Msg(msg)
}

// Warn uses fmt.Sprint to construct and log a message.
func (s *SugaredLogger) Warn(args ...interface{}) {
	e := s.logger.header(WarnLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	print(e.Context(s.context), args)
}

// Warnf uses fmt.Sprintf to log a templated message.
func (s *SugaredLogger) Warnf(template string, args ...interface{}) {
	e := s.logger.header(WarnLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	e.Context(s.context).Msgf(template, args...)
}

// Warnw logs a message with some additional context.
func (s *SugaredLogger) Warnw(msg string, keysAndValues ...interface{}) {
	e := s.logger.header(WarnLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	e.Context(s.context).keysAndValues(keysAndValues...).Msg(msg)
}

// Error uses fmt.Sprint to construct and log a message.
func (s *SugaredLogger) Error(args ...interface{}) {
	e := s.logger.header(ErrorLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	print(e.Context(s.context), args)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (s *SugaredLogger) Errorf(template string, args ...interface{}) {
	e := s.logger.header(ErrorLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	e.Context(s.context).Msgf(template, args...)
}

// Errorw logs a message with some additional context.
func (s *SugaredLogger) Errorw(msg string, keysAndValues ...interface{}) {
	e := s.logger.header(ErrorLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	e.Context(s.context).keysAndValues(keysAndValues...).Msg(msg)
}

// Fatal uses fmt.Sprint to construct and log a message.
func (s *SugaredLogger) Fatal(args ...interface{}) {
	e := s.logger.header(FatalLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	print(e.Context(s.context), args)
}

// Fatalf uses fmt.Sprintf to log a templated message.
func (s *SugaredLogger) Fatalf(template string, args ...interface{}) {
	e := s.logger.header(FatalLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	e.Context(s.context).Msgf(template, args...)
}

// Fatalw logs a message with some additional context.
func (s *SugaredLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	e := s.logger.header(FatalLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	e.Context(s.context).keysAndValues(keysAndValues...).Msg(msg)
}

// Panic uses fmt.Sprint to construct and log a message.
func (s *SugaredLogger) Panic(args ...interface{}) {
	e := s.logger.header(PanicLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	print(e.Context(s.context), args)
}

// Panicf uses fmt.Sprintf to log a templated message.
func (s *SugaredLogger) Panicf(template string, args ...interface{}) {
	e := s.logger.header(PanicLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	e.Context(s.context).Msgf(template, args...)
}

// Panicw logs a message with some additional context.
func (s *SugaredLogger) Panicw(msg string, keysAndValues ...interface{}) {
	e := s.logger.header(PanicLevel)
	if e == nil {
		return
	}
	if s.logger.Caller > 0 {
		e.caller(runtime.Caller(s.logger.Caller))
	}
	e.Context(s.context).keysAndValues(keysAndValues...).Msg(msg)
}

func print(e *Event, args []interface{}) {
	b := bbpool.Get().(*bb)
	b.Reset()

	fmt.Fprint(b, args...)
	e.Msg(*(*string)(unsafe.Pointer(&b.B)))

	if cap(b.B) <= bbcap {
		bbpool.Put(b)
	}
}
