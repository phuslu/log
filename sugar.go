package log

import (
	"fmt"
	"net"
	"runtime"
	"time"
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
	log(e.Context(s.context), keysAndValues)
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
	log(e.Context(s.context).Str("message", msg), keysAndValues)
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
	log(e.Context(s.context).Str("message", msg), keysAndValues)
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
	log(e.Context(s.context).Str("message", msg), keysAndValues)
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
	log(e.Context(s.context).Str("message", msg), keysAndValues)
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
	log(e.Context(s.context).Str("message", msg), keysAndValues)
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
	log(e.Context(s.context).Str("message", msg), keysAndValues)
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

func log(e *Event, keysAndValues []interface{}) {
	var key string
	for i, v := range keysAndValues {
		if i%2 == 0 {
			key, _ = v.(string)
			continue
		}
		if v == nil {
			e.key(key)
			e.buf = append(e.buf, "null"...)
			continue
		}
		switch v.(type) {
		case Context:
			e.Dict(key, v.(Context))
		case []time.Duration:
			e.Durs(key, v.([]time.Duration))
		case time.Duration:
			e.Dur(key, v.(time.Duration))
		case time.Time:
			e.Time(key, v.(time.Time))
		case net.HardwareAddr:
			e.MACAddr(key, v.(net.HardwareAddr))
		case net.IP:
			e.IPAddr(key, v.(net.IP))
		case net.IPNet:
			e.IPPrefix(key, v.(net.IPNet))
		case []bool:
			e.Bools(key, v.([]bool))
		case []byte:
			e.Bytes(key, v.([]byte))
		case []error:
			e.Errs(key, v.([]error))
		case []float32:
			e.Floats32(key, v.([]float32))
		case []float64:
			e.Floats64(key, v.([]float64))
		case []string:
			e.Strs(key, v.([]string))
		case string:
			e.Str(key, v.(string))
		case bool:
			e.Bool(key, v.(bool))
		case error:
			e.AnErr(key, v.(error))
		case float32:
			e.Float32(key, v.(float32))
		case float64:
			e.Float64(key, v.(float64))
		case int16:
			e.Int16(key, v.(int16))
		case int32:
			e.Int32(key, v.(int32))
		case int64:
			e.Int64(key, v.(int64))
		case int8:
			e.Int8(key, v.(int8))
		case int:
			e.Int(key, v.(int))
		case uint16:
			e.Uint16(key, v.(uint16))
		case uint32:
			e.Uint32(key, v.(uint32))
		case uint64:
			e.Uint64(key, v.(uint64))
		case uint8:
			e.Uint8(key, v.(uint8))
		case fmt.GoStringer:
			e.GoStringer(key, v.(fmt.GoStringer))
		case fmt.Stringer:
			e.Stringer(key, v.(fmt.Stringer))
		default:
			e.Interface(key, v)
		}
	}
	e.Msg("")
}
