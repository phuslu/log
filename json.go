package log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// DefaultLogger is the global logger.
var DefaultLogger = Logger{
	Level:      DebugLevel,
	Caller:     0,
	TimeField:  "",
	TimeFormat: "",
	HostField:  "",
	Writer:     &Writer{},
}

// A Logger represents an active logging object that generates lines of JSON output to an io.Writer.
type Logger struct {
	Level      Level
	Caller     int
	TimeField  string
	TimeFormat string
	HostField  string
	Writer     io.Writer
}

// Event represents a log event. It is instanced by one of the level method of Logger and finalized by the Msg or Msgf method.
type Event struct {
	buf   []byte
	write func(p []byte) (n int, err error)
	level Level
}

// Debug starts a new message with debug level.
func Debug() (e *Event) {
	e = DefaultLogger.header(DebugLevel)
	if e != nil && DefaultLogger.Caller > 0 {
		e.caller(runtime.Caller(DefaultLogger.Caller))
	}
	return
}

// Info starts a new message with info level.
func Info() (e *Event) {
	e = DefaultLogger.header(InfoLevel)
	if e != nil && DefaultLogger.Caller > 0 {
		e.caller(runtime.Caller(DefaultLogger.Caller))
	}
	return
}

// Warn starts a new message with warning level.
func Warn() (e *Event) {
	e = DefaultLogger.header(WarnLevel)
	if e != nil && DefaultLogger.Caller > 0 {
		e.caller(runtime.Caller(DefaultLogger.Caller))
	}
	return
}

// Error starts a new message with error level.
func Error() (e *Event) {
	e = DefaultLogger.header(ErrorLevel)
	if e != nil && DefaultLogger.Caller > 0 {
		e.caller(runtime.Caller(DefaultLogger.Caller))
	}
	return
}

// Fatal starts a new message with fatal level.
func Fatal() (e *Event) {
	e = DefaultLogger.header(FatalLevel)
	if e != nil && DefaultLogger.Caller > 0 {
		e.caller(runtime.Caller(DefaultLogger.Caller))
	}
	return
}

// Print sends a log event using debug level and no extra field. Arguments are handled in the manner of fmt.Print.
func Print(v ...interface{}) {
	e := DefaultLogger.header(DefaultLogger.Level)
	if e != nil && DefaultLogger.Caller > 0 {
		e.caller(runtime.Caller(DefaultLogger.Caller))
	}
	e.Msg(fmt.Sprint(v...))
}

// Printf sends a log event using debug level and no extra field. Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	e := DefaultLogger.header(DefaultLogger.Level)
	if e != nil && DefaultLogger.Caller > 0 {
		e.caller(runtime.Caller(DefaultLogger.Caller))
	}
	e.Msgf(format, v...)
}

// Debug starts a new message with debug level.
func (l Logger) Debug() (e *Event) {
	e = l.header(DebugLevel)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	return
}

// Info starts a new message with info level.
func (l Logger) Info() (e *Event) {
	e = l.header(InfoLevel)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	return
}

// Warn starts a new message with warning level.
func (l Logger) Warn() (e *Event) {
	e = l.header(WarnLevel)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	return
}

// Error starts a new message with error level.
func (l Logger) Error() (e *Event) {
	e = l.header(ErrorLevel)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	return
}

// Fatal starts a new message with fatal level.
func (l Logger) Fatal() (e *Event) {
	e = l.header(FatalLevel)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	return
}

// WithLevel starts a new message with level.
func (l Logger) WithLevel(level Level) (e *Event) {
	e = l.header(level)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	return
}

// SetLevel changes logger default level.
func (l *Logger) SetLevel(level Level) {
	atomic.StoreUint32((*uint32)(&l.Level), uint32(level))
	return
}

// Print sends a log event using debug level and no extra field. Arguments are handled in the manner of fmt.Print.
func (l Logger) Print(v ...interface{}) {
	e := l.header(l.Level)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	e.Msg(fmt.Sprint(v...))
}

// Printf sends a log event using debug level and no extra field. Arguments are handled in the manner of fmt.Printf.
func (l Logger) Printf(format string, v ...interface{}) {
	e := l.header(l.Level)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	e.Msgf(format, v...)
}

var epool = sync.Pool{
	New: func() interface{} {
		return new(Event)
	},
}

func (l Logger) header(level Level) (e *Event) {
	if uint32(level) < atomic.LoadUint32((*uint32)(&l.Level)) {
		return
	}
	e = epool.Get().(*Event)
	e.buf = e.buf[:0]
	e.level = level
	e.write = l.Writer.Write
	// time
	now := timeNow()
	if l.TimeField == "" {
		e.buf = append(e.buf, "{\"time\":"...)
	} else {
		e.buf = append(e.buf, '{', '"')
		e.buf = append(e.buf, l.TimeField...)
		e.buf = append(e.buf, '"', ':')
	}
	if l.TimeFormat == "" {
		e.time(now)
	} else {
		e.buf = append(e.buf, '"')
		e.buf = now.AppendFormat(e.buf, l.TimeFormat)
		e.buf = append(e.buf, '"')
	}
	// level
	switch level {
	case DebugLevel:
		e.buf = append(e.buf, ",\"level\":\"debug\""...)
	case InfoLevel:
		e.buf = append(e.buf, ",\"level\":\"info\""...)
	case WarnLevel:
		e.buf = append(e.buf, ",\"level\":\"warn\""...)
	case ErrorLevel:
		e.buf = append(e.buf, ",\"level\":\"error\""...)
	case FatalLevel:
		e.buf = append(e.buf, ",\"level\":\"fatal\""...)
	}
	// hostname
	if l.HostField != "" {
		e.buf = append(e.buf, ',', '"')
		e.buf = append(e.buf, l.HostField...)
		e.buf = append(e.buf, '"', ':', '"')
		e.buf = append(e.buf, hostname...)
		e.buf = append(e.buf, '"')
	}
	return
}

// Time append append t formated as string using time.RFC3339Nano.
func (e *Event) Time(key string, t time.Time) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = append(e.buf, '"')
	e.buf = t.AppendFormat(e.buf, time.RFC3339Nano)
	e.buf = append(e.buf, '"')
	return e
}

// TimeFormat append append t formated as string using timefmt.
func (e *Event) TimeFormat(key string, timefmt string, t time.Time) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = append(e.buf, '"')
	e.buf = t.AppendFormat(e.buf, timefmt)
	e.buf = append(e.buf, '"')
	return e
}

// Timestamp adds the current time as UNIX timestamp
func (e *Event) Timestamp() *Event {
	if e == nil {
		return nil
	}
	e.key("timestamp")
	e.buf = strconv.AppendInt(e.buf, timeNow().Unix(), 10)
	return e
}

// Bool append append the val as a bool to the event.
func (e *Event) Bool(key string, b bool) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = strconv.AppendBool(e.buf, b)
	return e
}

// Bools adds the field key with val as a []bool to the event.
func (e *Event) Bools(key string, b []bool) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = append(e.buf, '[')
	for i, a := range b {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendBool(e.buf, a)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Dur adds the field key with duration d to the event.
func (e *Event) Dur(key string, d time.Duration) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = append(e.buf, '"')
	e.buf = append(e.buf, d.String()...)
	e.buf = append(e.buf, '"')
	return e
}

// Durs adds the field key with val as a []time.Duration to the event.
func (e *Event) Durs(key string, d []time.Duration) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = append(e.buf, '[')
	for i, a := range d {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = append(e.buf, '"')
		e.buf = append(e.buf, a.String()...)
		e.buf = append(e.buf, '"')
	}
	e.buf = append(e.buf, ']')
	return e
}

// Err adds the field "error" with serialized err to the event.
func (e *Event) Err(err error) *Event {
	if e == nil {
		return nil
	}
	if err == nil {
		e.buf = append(e.buf, ",\"error\":null"...)
	} else {
		e.buf = append(e.buf, ",\"error\":"...)
		e.string(err.Error())
	}
	return e
}

// Errs adds the field key with errs as an array of serialized errors to the event.
func (e *Event) Errs(key string, errs []error) *Event {
	if e == nil {
		return nil
	}

	e.key(key)
	e.buf = append(e.buf, '[')
	for i, err := range errs {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		if err == nil {
			e.buf = append(e.buf, "null"...)
		} else {
			e.string(err.Error())
		}
	}
	e.buf = append(e.buf, ']')
	return e
}

// Float64 adds the field key with f as a float64 to the event.
func (e *Event) Float64(key string, f float64) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = strconv.AppendFloat(e.buf, f, 'f', -1, 64)
	return e
}

// Floats64 adds the field key with f as a []float64 to the event.
func (e *Event) Floats64(key string, f []float64) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = append(e.buf, '[')
	for i, a := range f {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendFloat(e.buf, a, 'f', -1, 64)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Floats32 adds the field key with f as a []float32 to the event.
func (e *Event) Floats32(key string, f []float32) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = append(e.buf, '[')
	for i, a := range f {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendFloat(e.buf, float64(a), 'f', -1, 64)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Int64 adds the field key with i as a int64 to the event.
func (e *Event) Int64(key string, i int64) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = strconv.AppendInt(e.buf, i, 10)
	return e
}

// Uint64 adds the field key with i as a uint64 to the event.
func (e *Event) Uint64(key string, i uint64) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = strconv.AppendUint(e.buf, i, 10)
	return e
}

// Float32 adds the field key with f as a float32 to the event.
func (e *Event) Float32(key string, f float32) *Event {
	return e.Float64(key, float64(f))
}

// Int adds the field key with i as a int to the event.
func (e *Event) Int(key string, i int) *Event {
	return e.Int64(key, int64(i))
}

// Int32 adds the field key with i as a int32 to the event.
func (e *Event) Int32(key string, i int32) *Event {
	return e.Int64(key, int64(i))
}

// Int16 adds the field key with i as a int16 to the event.
func (e *Event) Int16(key string, i int16) *Event {
	return e.Int64(key, int64(i))
}

// Int8 adds the field key with i as a int8 to the event.
func (e *Event) Int8(key string, i int8) *Event {
	return e.Int64(key, int64(i))
}

// Uint32 adds the field key with i as a uint32 to the event.
func (e *Event) Uint32(key string, i uint32) *Event {
	return e.Uint64(key, uint64(i))
}

// Uint16 adds the field key with i as a uint16 to the event.
func (e *Event) Uint16(key string, i uint16) *Event {
	return e.Uint64(key, uint64(i))
}

// Uint8 adds the field key with i as a uint8 to the event.
func (e *Event) Uint8(key string, i uint8) *Event {
	return e.Uint64(key, uint64(i))
}

// RawJSON adds already encoded JSON to the log line under key.
func (e *Event) RawJSON(key string, b []byte) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = append(e.buf, b...)
	return e
}

// Str adds the field key with val as a string to the event.
func (e *Event) Str(key string, val string) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.string(val)
	return e
}

// Strs adds the field key with vals as a []string to the event.
func (e *Event) Strs(key string, vals []string) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = append(e.buf, '[')
	for i, val := range vals {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.string(val)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Bytes adds the field key with val as a string to the event.
func (e *Event) Bytes(key string, val []byte) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.bytes(val)
	return e
}

// Hex adds the field key with val as a hex string to the event.
func (e *Event) Hex(key string, val []byte) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = append(e.buf, '"')
	for _, v := range val {
		e.buf = append(e.buf, hex[v>>4], hex[v&0x0f])
	}
	e.buf = append(e.buf, '"')
	return e
}

// Interface adds the field key with i marshaled using reflection.
func (e *Event) Interface(key string, i interface{}) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	marshaled, err := json.Marshal(i)
	if err != nil {
		e.string("marshaling error: " + err.Error())
	} else {
		e.bytes(marshaled)
	}
	return e
}

// Caller adds the file:line of the "caller" key.
func (e *Event) Caller() *Event {
	if e == nil {
		return nil
	}
	e.caller(runtime.Caller(DefaultLogger.Caller))
	return e
}

// Enabled return false if the event is going to be filtered out by log level.
func (e *Event) Enabled() bool {
	return e != nil
}

// Discard disables the event so Msg(f) won't print it.
func (e *Event) Discard() *Event {
	if e == nil {
		return e
	}
	epool.Put(e)
	return nil
}

// Msg sends the event with msg added as the message field if not empty.
func (e *Event) Msg(msg string) {
	if e == nil {
		return
	}
	if msg != "" {
		e.buf = append(e.buf, ",\"message\":"...)
		e.string(msg)
	}
	e.buf = append(e.buf, '}', '\n')
	e.write(e.buf)
	if e.level == FatalLevel {
		e.write(stacks(false))
		e.write(stacks(true))
		os.Exit(255)
	}
	epool.Put(e)
}

// Msgf sends the event with formatted msg added as the message field if not empty.
func (e *Event) Msgf(format string, v ...interface{}) {
	e.Msg(fmt.Sprintf(format, v...))
}

func (e *Event) key(key string) {
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
}

func (e *Event) caller(_ uintptr, file string, line int, _ bool) {
	if i := strings.LastIndex(file, "/"); i >= 0 {
		file = file[i+1:]
	}
	e.buf = append(e.buf, ",\"caller\":\""...)
	e.buf = append(e.buf, file...)
	e.buf = append(e.buf, ':')
	e.buf = strconv.AppendInt(e.buf, int64(line), 10)
	e.buf = append(e.buf, '"')
}

const timebuf = "\"2006-01-02T15:04:05.999Z\""

func (e *Event) time(now time.Time) {
	now = now.UTC()
	n := len(e.buf)
	if n+len(timebuf) < cap(e.buf) {
		e.buf = e.buf[:n+len(timebuf)]
	} else {
		e.buf = append(e.buf, timebuf...)
	}
	var a, b int
	// milli second
	e.buf[n+25] = '"'
	e.buf[n+24] = 'Z'
	a = now.Nanosecond() / 1000000
	b = a / 10
	e.buf[n+23] = byte('0' + a - 10*b)
	a = b
	b = a / 10
	e.buf[n+22] = byte('0' + a - 10*b)
	e.buf[n+21] = byte('0' + b)
	e.buf[n+20] = '.'
	// date time
	year, month, day := now.Date()
	hour, minute, second := now.Clock()
	// year
	a = year
	b = a / 10
	e.buf[n+4] = byte('0' + a - 10*b)
	a = b
	b = a / 10
	e.buf[n+3] = byte('0' + a - 10*b)
	a = b
	b = a / 10
	e.buf[n+2] = byte('0' + a - 10*b)
	e.buf[n+1] = byte('0' + b)
	e.buf[n] = '"'
	// month
	a = int(month)
	b = a / 10
	e.buf[n+7] = byte('0' + a - 10*b)
	e.buf[n+6] = byte('0' + b)
	e.buf[n+5] = '-'
	// day
	a = day
	b = a / 10
	e.buf[n+10] = byte('0' + a - 10*b)
	e.buf[n+9] = byte('0' + b)
	e.buf[n+8] = '-'
	// hour
	a = hour
	b = a / 10
	e.buf[n+13] = byte('0' + a - 10*b)
	e.buf[n+12] = byte('0' + b)
	e.buf[n+11] = 'T'
	// minute
	a = minute
	b = a / 10
	e.buf[n+16] = byte('0' + a - 10*b)
	e.buf[n+15] = byte('0' + b)
	e.buf[n+14] = ':'
	// second
	a = second
	b = a / 10
	e.buf[n+19] = byte('0' + a - 10*b)
	e.buf[n+18] = byte('0' + b)
	e.buf[n+17] = ':'
}

var hex = "0123456789abcdef"

// refer to https://github.com/valyala/quicktemplate/blob/master/jsonstring.go
func (e *Event) string(s string) {
	if n := len(s); n > 24 {
		var needEscape bool
		for i := 0; i < n; i++ {
			switch s[i] {
			case '"', '\\', '\n', '\r', '\t', '\f', '\b', '<', '\'', 0:
				needEscape = true
				break
			}
		}
		// fast path - nothing to escape
		if !needEscape {
			e.buf = append(e.buf, '"')
			e.buf = append(e.buf, s...)
			e.buf = append(e.buf, '"')
			return
		}
	}

	// slow path
	e.buf = append(e.buf, '"')
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	b := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{Data: sh.Data, Len: sh.Len, Cap: sh.Len}))
	j := 0
	n := len(b)
	if n > 0 {
		// Hint the compiler to remove bounds checks in the loop below.
		_ = b[n-1]
	}
	for i := 0; i < n; i++ {
		switch b[i] {
		case '"':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', '"')
			j = i + 1
		case '\\':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', '\\')
			j = i + 1
		case '\n':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 'n')
			j = i + 1
		case '\r':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 'r')
			j = i + 1
		case '\t':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 't')
			j = i + 1
		case '\f':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '0', 'c')
			j = i + 1
		case '\b':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '0', '8')
			j = i + 1
		case '<':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '3', 'c')
			j = i + 1
		case '\'':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '2', '7')
			j = i + 1
		case 0:
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '0', '0')
			j = i + 1
		}
	}
	e.buf = append(e.buf, b[j:]...)
	e.buf = append(e.buf, '"')
}

func (e *Event) bytes(b []byte) {
	n := len(b)
	if n > 24 {
		var needEscape bool
		for i := 0; i < n; i++ {
			switch b[i] {
			case '"', '\\', '\n', '\r', '\t', '\f', '\b', '<', '\'', 0:
				needEscape = true
				break
			}
		}
		// fast path - nothing to escape
		if !needEscape {
			e.buf = append(e.buf, '"')
			e.buf = append(e.buf, b...)
			e.buf = append(e.buf, '"')
			return
		}
	}

	// slow path
	e.buf = append(e.buf, '"')
	j := 0
	if n > 0 {
		// Hint the compiler to remove bounds checks in the loop below.
		_ = b[n-1]
	}
	for i := 0; i < n; i++ {
		switch b[i] {
		case '"':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', '"')
			j = i + 1
		case '\\':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', '\\')
			j = i + 1
		case '\n':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 'n')
			j = i + 1
		case '\r':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 'r')
			j = i + 1
		case '\t':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 't')
			j = i + 1
		case '\f':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '0', 'c')
			j = i + 1
		case '\b':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '0', '8')
			j = i + 1
		case '<':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '3', 'c')
			j = i + 1
		case '\'':
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '2', '7')
			j = i + 1
		case 0:
			e.buf = append(e.buf, b[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '0', '0')
			j = i + 1
		}
	}
	e.buf = append(e.buf, b[j:]...)
	e.buf = append(e.buf, '"')
}

// stacks is a wrapper for runtime.Stack that attempts to recover the data for all goroutines.
func stacks(all bool) []byte {
	// We don't know how big the traces are, so grow a few times if they don't fit. Start large, though.
	n := 10000
	if all {
		n = 100000
	}
	var trace []byte
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, all)
		if nbytes < len(trace) {
			return trace[:nbytes]
		}
		n *= 2
	}
	return trace
}

var _ io.Writer = (*LevelWriter)(nil)

// LevelWriter defines as interface a writer may implement in order to receive level information with payload.
type LevelWriter struct {
	Logger Logger
	Level  Level
}

func (w LevelWriter) Write(p []byte) (int, error) {
	e := w.Logger.header(w.Level)
	if e != nil && w.Logger.Caller > 0 {
		e.caller(runtime.Caller(w.Logger.Caller))
	}
	e.Msg(*(*string)(unsafe.Pointer(&p)))
	return len(p), nil
}

// Fastrandn returns a pseudorandom uint32 in [0,n).
//go:noescape
//go:linkname Fastrandn runtime.fastrandn
func Fastrandn(x uint32) uint32
