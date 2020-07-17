package log

import (
	"fmt"
	"net"
	"runtime"
	"time"
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
	level   Level
	context Context
}

// Sugar wraps the Logger to provide a more ergonomic, but slightly slower,
// API. Sugaring a Logger is quite inexpensive, so it's reasonable for a
// single application to use both Loggers and SugaredLoggers, converting
// between them on the boundaries of performance-sensitive code.
func (l *Logger) Sugar(level Level, context Context) (logger *SugaredLogger) {
	logger = &SugaredLogger{
		logger:  *l,
		level:   level,
		context: context,
	}
	return
}

// Print sends a log event without extra field. Arguments are handled in the manner of fmt.Print.
func (l *SugaredLogger) Print(v ...interface{}) {
	e := l.logger.header(l.level)
	if e == nil {
		return
	}
	if l.logger.Caller > 0 {
		e.caller(runtime.Caller(l.logger.Caller))
	}
	e.print(v...)
}

// Println sends a log event without extra field. Arguments are handled in the manner of fmt.Print.
func (l *SugaredLogger) Println(v ...interface{}) {
	e := l.logger.header(l.level)
	if e == nil {
		return
	}
	if l.logger.Caller > 0 {
		e.caller(runtime.Caller(l.logger.Caller))
	}
	e.print(v...)
}

// Printf sends a log event without extra field. Arguments are handled in the manner of fmt.Printf.
func (l *SugaredLogger) Printf(format string, v ...interface{}) {
	e := l.logger.header(l.level)
	if e == nil {
		return
	}
	if l.logger.Caller > 0 {
		e.caller(runtime.Caller(l.logger.Caller))
	}
	e.Msgf(format, v...)
}

// Log sends a log event without extra field. Arguments are handled in the manner of fmt.Printf.
func (l *SugaredLogger) Log(keyvals ...interface{}) error {
	e := l.logger.header(l.level)
	if e == nil {
		return nil
	}
	if l.logger.Caller > 0 {
		e.caller(runtime.Caller(l.logger.Caller))
	}
	if l.context != nil {
		e.buf = append(e.buf, l.context...)
	}
	var key, msg string
	if len(keyvals)%2 == 1 {
		msg, _ = keyvals[0].(string)
		keyvals = keyvals[1:]
	}
	for i, v := range keyvals {
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
	e.Msg(msg)
	return nil
}
