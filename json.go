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
	"time"
	"unsafe"
)

var DefaultLogger = Logger{
	Level:      DebugLevel,
	Caller:     false,
	TimeField:  "",
	TimeFormat: "",
	Writer:     &Writer{},
}

type Logger struct {
	Level      Level
	Caller     bool
	TimeField  string
	TimeFormat string
	Writer     io.Writer
}

type Event struct {
	buf        []byte
	level      Level
	timeFormat string
	write      func(p []byte) (n int, err error)
}

func Debug() *Event {
	return DefaultLogger.withLevel(DebugLevel)
}

func Info() *Event {
	return DefaultLogger.withLevel(InfoLevel)
}

func Warn() *Event {
	return DefaultLogger.withLevel(WarnLevel)
}

func Error() *Event {
	return DefaultLogger.withLevel(ErrorLevel)
}

func Fatal() *Event {
	return DefaultLogger.withLevel(FatalLevel)
}

func Print(v ...interface{}) {
	DefaultLogger.Print(v...)
}

func Printf(format string, v ...interface{}) {
	DefaultLogger.Printf(format, v...)
}

func (l Logger) Debug() *Event {
	return l.withLevel(DebugLevel)
}

func (l Logger) Info() *Event {
	return l.withLevel(InfoLevel)
}

func (l Logger) Warn() *Event {
	return l.withLevel(WarnLevel)
}

func (l Logger) Error() *Event {
	return l.withLevel(ErrorLevel)
}

func (l Logger) Fatal() *Event {
	return l.withLevel(FatalLevel)
}

func (l Logger) Print(v ...interface{}) {
	l.withLevel(l.Level).Msg(fmt.Sprint(v...))
}

func (l Logger) Printf(format string, v ...interface{}) {
	l.withLevel(l.Level).Msgf(format, v...)
}

var epool = sync.Pool{
	New: func() interface{} {
		return new(Event)
	},
}

func (l Logger) withLevel(level Level) (e *Event) {
	if level < l.Level {
		return
	}
	e = epool.Get().(*Event)
	e.buf = e.buf[:0]
	e.level = level
	e.timeFormat = l.TimeFormat
	e.write = l.Writer.Write
	// time
	now := timeNow()
	if l.TimeField == "" {
		e.buf = append(e.buf, "{\"time\":"...)
	} else {
		e.key('{', l.TimeField)
	}
	if e.timeFormat == "" {
		e.time(now)
	} else {
		e.buf = append(e.buf, '"')
		e.buf = now.AppendFormat(e.buf, e.timeFormat)
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
	// caller
	if l.Caller {
		_, file, line, ok := runtime.Caller(2)
		if !ok {
			file = "???"
			line = 1
		} else {
			if i := strings.LastIndex(file, "/"); i >= 0 {
				file = file[i+1:]
			}
		}
		if line < 0 {
			line = 0
		}
		e.buf = append(e.buf, ",\"caller\":\""...)
		e.buf = append(e.buf, file...)
		e.buf = append(e.buf, ':')
		e.buf = strconv.AppendInt(e.buf, int64(line), 10)
		e.buf = append(e.buf, '"')
	}
	return
}

func (e *Event) Time(key string, t time.Time) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
	switch {
	case e.timeFormat != "":
		e.buf = append(e.buf, '"')
		e.buf = t.AppendFormat(e.buf, e.timeFormat)
		e.buf = append(e.buf, '"')
	default:
		e.time(t)
	}
	return e
}

func (e *Event) Timestamp() *Event {
	if e == nil {
		return nil
	}
	e.key(',', "timestamp")
	e.buf = strconv.AppendInt(e.buf, timeNow().Unix(), 10)
	return e
}

func (e *Event) Bool(key string, b bool) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
	e.buf = strconv.AppendBool(e.buf, b)
	return e
}

func (e *Event) Bools(key string, b []bool) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
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

func (e *Event) Dur(key string, d time.Duration) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
	e.buf = append(e.buf, '"')
	e.buf = append(e.buf, d.String()...)
	e.buf = append(e.buf, '"')
	return e
}

func (e *Event) Durs(key string, d []time.Duration) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
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

func (e *Event) Errs(key string, errs []error) *Event {
	if e == nil {
		return nil
	}

	e.key(',', key)
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

func (e *Event) Float64(key string, f float64) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
	e.buf = strconv.AppendFloat(e.buf, f, 'f', -1, 64)
	return e
}

func (e *Event) Floats64(key string, f []float64) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
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

func (e *Event) Floats32(key string, f []float32) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
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

func (e *Event) Int64(key string, i int64) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
	e.buf = strconv.AppendInt(e.buf, i, 10)
	return e
}

func (e *Event) Uint64(key string, i uint64) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
	e.buf = strconv.AppendUint(e.buf, i, 10)
	return e
}

func (e *Event) Float32(key string, f float32) *Event {
	return e.Float64(key, float64(f))
}

func (e *Event) Int(key string, i int) *Event {
	return e.Int64(key, int64(i))
}

func (e *Event) Int32(key string, i int32) *Event {
	return e.Int64(key, int64(i))
}

func (e *Event) Int16(key string, i int16) *Event {
	return e.Int64(key, int64(i))
}

func (e *Event) Int8(key string, i int8) *Event {
	return e.Int64(key, int64(i))
}

func (e *Event) Uint32(key string, i uint32) *Event {
	return e.Uint64(key, uint64(i))
}

func (e *Event) Uint16(key string, i uint16) *Event {
	return e.Uint64(key, uint64(i))
}

func (e *Event) Uint8(key string, i uint8) *Event {
	return e.Uint64(key, uint64(i))
}

func (e *Event) RawJSON(key string, b []byte) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
	e.buf = append(e.buf, b...)
	return e
}

func (e *Event) Str(key string, val string) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
	e.string(val)
	return e
}

func (e *Event) Strs(key string, vals []string) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
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

func (e *Event) Bytes(key string, val []byte) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
	e.bytes(val)
	return e
}

func (e *Event) Hex(key string, val []byte) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
	e.buf = append(e.buf, '"')
	for _, v := range val {
		e.buf = append(e.buf, hex[v>>4], hex[v&0x0f])
	}
	e.buf = append(e.buf, '"')
	return e
}

func (e *Event) Interface(key string, i interface{}) *Event {
	if e == nil {
		return nil
	}
	e.key(',', key)
	marshaled, err := json.Marshal(i)
	if err != nil {
		e.string("marshaling error: " + err.Error())
	} else {
		e.bytes(marshaled)
	}
	return e
}

func (e *Event) Caller() *Event {
	if e == nil {
		return nil
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 1
	} else {
		if i := strings.LastIndex(file, "/"); i >= 0 {
			file = file[i+1:]
		}
	}
	if line < 0 {
		line = 0
	}
	e.buf = append(e.buf, ",\"caller\":\""...)
	e.buf = append(e.buf, file...)
	e.buf = append(e.buf, ':')
	e.buf = strconv.AppendInt(e.buf, int64(line), 10)
	e.buf = append(e.buf, '"')
	return e
}

func (e *Event) Send() {
	e.Msg("")
}

func (e *Event) Enabled() bool {
	return e != nil
}

func (e *Event) Discard() *Event {
	if e == nil {
		return e
	}
	epool.Put(e)
	return nil
}

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

func (e *Event) Msgf(format string, v ...interface{}) {
	e.Msg(fmt.Sprintf(format, v...))
}

func (e *Event) key(b byte, key string) {
	e.buf = append(e.buf, b, '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
}

var timebuf []byte = []byte("\"2006-01-02T15:04:05.999Z\"")

func (e *Event) time(now time.Time) {
	now = now.UTC()
	n := len(e.buf)
	if n+26 < cap(e.buf) {
		e.buf = e.buf[:n+26]
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
	// year
	a = now.Year()
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
	a = int(now.Month())
	b = a / 10
	e.buf[n+7] = byte('0' + a - 10*b)
	e.buf[n+6] = byte('0' + b)
	e.buf[n+5] = '-'
	// day
	a = now.Day()
	b = a / 10
	e.buf[n+10] = byte('0' + a - 10*b)
	e.buf[n+9] = byte('0' + b)
	e.buf[n+8] = '-'
	// hour
	a = now.Hour()
	b = a / 10
	e.buf[n+13] = byte('0' + a - 10*b)
	e.buf[n+12] = byte('0' + b)
	e.buf[n+11] = 'T'
	// minute
	a = now.Minute()
	b = a / 10
	e.buf[n+16] = byte('0' + a - 10*b)
	e.buf[n+15] = byte('0' + b)
	e.buf[n+14] = ':'
	// second
	a = now.Second()
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
