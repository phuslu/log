package log

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
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
	Writer:     IOWriter{os.Stderr},
}

// Entry represents a log entry. It is instanced by one of the level method of Logger and finalized by the Msg or Msgf method.
type Entry struct {
	buf   []byte
	Level Level
	w     Writer
}

// Writer defines an entry writer interface.
type Writer interface {
	WriteEntry(*Entry) (int, error)
}

// IOWriter wraps an io.Writer to Writer.
type IOWriter struct {
	io.Writer
}

// WriteEntry implements Writer.
func (w IOWriter) WriteEntry(e *Entry) (n int, err error) {
	return w.Writer.Write(e.buf)
}

// IOWriteCloser wraps an io.IOWriteCloser to Writer.
type IOWriteCloser struct {
	io.WriteCloser
}

// WriteEntry implements Writer.
func (w IOWriteCloser) WriteEntry(e *Entry) (n int, err error) {
	return w.WriteCloser.Write(e.buf)
}

// Close implements Writer.
func (w IOWriteCloser) Close() (err error) {
	return w.WriteCloser.Close()
}

// ObjectMarshaler provides a strongly-typed and encoding-agnostic interface
// to be implemented by types used with Entry's Object methods.
type ObjectMarshaler interface {
	MarshalObject(e *Entry)
}

// A Logger represents an active logging object that generates lines of JSON output to an io.Writer.
type Logger struct {
	// Level defines log levels.
	Level Level

	// Caller determines if adds the file:line of the "caller" key.
	// If Caller is negative, adds the full /path/to/file:line of the "caller" key.
	Caller int

	// TimeField defines the time filed name in output.  It uses "time" in if empty.
	TimeField string

	// TimeFormat specifies the time format in output. It uses time.RFC3339 with milliseconds if empty.
	// If set with `TimeFormatUnix`, `TimeFormatUnixMs`, times are formated as UNIX timestamp.
	TimeFormat string

	// Context specifies an optional context of logger.
	Context Context

	// Writer specifies the writer of output. It uses a wrapped os.Stderr Writer in if empty.
	Writer Writer
}

// TimeFormatUnix defines a time format that makes time fields to be
// serialized as Unix timestamp integers.
const TimeFormatUnix = "\x01"

// TimeFormatUnixMs defines a time format that makes time fields to be
// serialized as Unix timestamp integers in milliseconds.
const TimeFormatUnixMs = "\x02"

// TimeFormatUnixWithMs defines a time format that makes time fields to be
// serialized as Unix timestamp timestamp floats.
const TimeFormatUnixWithMs = "\x03"

// Trace starts a new message with trace level.
func Trace() (e *Entry) {
	if DefaultLogger.silent(TraceLevel) {
		return nil
	}
	e = DefaultLogger.header(TraceLevel)
	if caller, full := DefaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Debug starts a new message with debug level.
func Debug() (e *Entry) {
	if DefaultLogger.silent(DebugLevel) {
		return nil
	}
	e = DefaultLogger.header(DebugLevel)
	if caller, full := DefaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Info starts a new message with info level.
func Info() (e *Entry) {
	if DefaultLogger.silent(InfoLevel) {
		return nil
	}
	e = DefaultLogger.header(InfoLevel)
	if caller, full := DefaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Warn starts a new message with warning level.
func Warn() (e *Entry) {
	if DefaultLogger.silent(WarnLevel) {
		return nil
	}
	e = DefaultLogger.header(WarnLevel)
	if caller, full := DefaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Error starts a new message with error level.
func Error() (e *Entry) {
	if DefaultLogger.silent(ErrorLevel) {
		return nil
	}
	e = DefaultLogger.header(ErrorLevel)
	if caller, full := DefaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Fatal starts a new message with fatal level.
func Fatal() (e *Entry) {
	if DefaultLogger.silent(FatalLevel) {
		return nil
	}
	e = DefaultLogger.header(FatalLevel)
	if caller, full := DefaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Panic starts a new message with panic level.
func Panic() (e *Entry) {
	if DefaultLogger.silent(PanicLevel) {
		return nil
	}
	e = DefaultLogger.header(PanicLevel)
	if caller, full := DefaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Printf sends a log entry without extra field. Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	e := DefaultLogger.header(noLevel)
	if caller, full := DefaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	e.Msgf(format, v...)
}

// Trace starts a new message with trace level.
func (l *Logger) Trace() (e *Entry) {
	if l.silent(TraceLevel) {
		return nil
	}
	e = l.header(TraceLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Debug starts a new message with debug level.
func (l *Logger) Debug() (e *Entry) {
	if l.silent(DebugLevel) {
		return nil
	}
	e = l.header(DebugLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Info starts a new message with info level.
func (l *Logger) Info() (e *Entry) {
	if l.silent(InfoLevel) {
		return nil
	}
	e = l.header(InfoLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Warn starts a new message with warning level.
func (l *Logger) Warn() (e *Entry) {
	if l.silent(WarnLevel) {
		return nil
	}
	e = l.header(WarnLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Error starts a new message with error level.
func (l *Logger) Error() (e *Entry) {
	if l.silent(ErrorLevel) {
		return nil
	}
	e = l.header(ErrorLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Fatal starts a new message with fatal level.
func (l *Logger) Fatal() (e *Entry) {
	if l.silent(FatalLevel) {
		return nil
	}
	e = l.header(FatalLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Panic starts a new message with panic level.
func (l *Logger) Panic() (e *Entry) {
	if l.silent(PanicLevel) {
		return nil
	}
	e = l.header(PanicLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Log starts a new message with no level.
func (l *Logger) Log() (e *Entry) {
	e = l.header(noLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// WithLevel starts a new message with level.
func (l *Logger) WithLevel(level Level) (e *Entry) {
	if l.silent(level) {
		return nil
	}
	e = l.header(level)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// Err starts a new message with error level with err as a field if not nil or with info level if err is nil.
func (l *Logger) Err(err error) (e *Entry) {
	var level = InfoLevel
	if err != nil {
		level = ErrorLevel
	}
	if l.silent(level) {
		return nil
	}
	e = l.header(level)
	if e == nil {
		return nil
	}
	if level == ErrorLevel {
		e = e.Err(err)
	}
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	return
}

// SetLevel changes logger default level.
func (l *Logger) SetLevel(level Level) {
	atomic.StoreUint32((*uint32)(&l.Level), uint32(level))
}

// Printf sends a log entry without extra field. Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Printf(format string, v ...interface{}) {
	e := l.header(noLevel)
	if e != nil {
		if caller, full := l.Caller, false; caller != 0 {
			if caller < 0 {
				caller, full = -caller, true
			}
			var rpc [1]uintptr
			e.caller(callers(caller, rpc[:]), rpc[:], full)
		}
	}
	e.Msgf(format, v...)
}

var epool = sync.Pool{
	New: func() interface{} {
		return &Entry{
			buf: make([]byte, 0, 1024),
		}
	},
}

const bbcap = 1 << 16

const smallsString = "00010203040506070809" +
	"10111213141516171819" +
	"20212223242526272829" +
	"30313233343536373839" +
	"40414243444546474849" +
	"50515253545556575859" +
	"60616263646566676869" +
	"70717273747576777879" +
	"80818283848586878889" +
	"90919293949596979899"

var timeNow = time.Now
var timeOffset, timeZone = func() (int64, string) {
	now := timeNow()
	_, n := now.Zone()
	s := now.Format("Z07:00")
	return int64(n), s
}()

func (l *Logger) silent(level Level) bool {
	return uint32(level) < atomic.LoadUint32((*uint32)(&l.Level))
}

func (l *Logger) header(level Level) *Entry {
	e := epool.Get().(*Entry)
	e.buf = e.buf[:0]
	e.Level = level
	if l.Writer != nil {
		e.w = l.Writer
	} else {
		e.w = IOWriter{os.Stderr}
	}
	// time
	if l.TimeField == "" {
		e.buf = append(e.buf, "{\"time\":"...)
	} else {
		e.buf = append(e.buf, '{', '"')
		e.buf = append(e.buf, l.TimeField...)
		e.buf = append(e.buf, '"', ':')
	}
	switch l.TimeFormat {
	case "":
		sec, nsec, _ := now()
		var tmp [32]byte
		var buf []byte
		if timeOffset == 0 {
			// "2006-01-02T15:04:05.999Z"
			tmp[25] = '"'
			tmp[24] = 'Z'
			buf = tmp[:26]
		} else {
			// "2006-01-02T15:04:05.999Z07:00"
			tmp[30] = '"'
			tmp[29] = timeZone[5]
			tmp[28] = timeZone[4]
			tmp[27] = timeZone[3]
			tmp[26] = timeZone[2]
			tmp[25] = timeZone[1]
			tmp[24] = timeZone[0]
			buf = tmp[:31]
		}
		// date time
		sec += 9223372028715321600 + timeOffset // unixToInternal + internalToAbsolute + timeOffset
		year, month, day, _ := absDate(uint64(sec), true)
		hour, minute, second := absClock(uint64(sec))
		// year
		a := year / 100 * 2
		b := year % 100 * 2
		tmp[0] = '"'
		tmp[1] = smallsString[a]
		tmp[2] = smallsString[a+1]
		tmp[3] = smallsString[b]
		tmp[4] = smallsString[b+1]
		// month
		month *= 2
		tmp[5] = '-'
		tmp[6] = smallsString[month]
		tmp[7] = smallsString[month+1]
		// day
		day *= 2
		tmp[8] = '-'
		tmp[9] = smallsString[day]
		tmp[10] = smallsString[day+1]
		// hour
		hour *= 2
		tmp[11] = 'T'
		tmp[12] = smallsString[hour]
		tmp[13] = smallsString[hour+1]
		// minute
		minute *= 2
		tmp[14] = ':'
		tmp[15] = smallsString[minute]
		tmp[16] = smallsString[minute+1]
		// second
		second *= 2
		tmp[17] = ':'
		tmp[18] = smallsString[second]
		tmp[19] = smallsString[second+1]
		// milli seconds
		a = int(nsec) / 1000000
		b = a % 100 * 2
		tmp[20] = '.'
		tmp[21] = byte('0' + a/100)
		tmp[22] = smallsString[b]
		tmp[23] = smallsString[b+1]
		// append to e.buf
		e.buf = append(e.buf, buf...)
	case TimeFormatUnix:
		sec, _, _ := now()
		// 1595759807
		var tmp [10]byte
		// seconds
		b := sec % 100 * 2
		sec /= 100
		tmp[9] = smallsString[b+1]
		tmp[8] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[7] = smallsString[b+1]
		tmp[6] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[5] = smallsString[b+1]
		tmp[4] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[3] = smallsString[b+1]
		tmp[2] = smallsString[b]
		b = sec % 100 * 2
		tmp[1] = smallsString[b+1]
		tmp[0] = smallsString[b]
		// append to e.buf
		e.buf = append(e.buf, tmp[:]...)
	case TimeFormatUnixMs:
		sec, nsec, _ := now()
		// 1595759807105
		var tmp [13]byte
		// milli seconds
		a := int64(nsec) / 1000000
		b := a % 100 * 2
		tmp[12] = smallsString[b+1]
		tmp[11] = smallsString[b]
		tmp[10] = byte('0' + a/100)
		// seconds
		b = sec % 100 * 2
		sec /= 100
		tmp[9] = smallsString[b+1]
		tmp[8] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[7] = smallsString[b+1]
		tmp[6] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[5] = smallsString[b+1]
		tmp[4] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[3] = smallsString[b+1]
		tmp[2] = smallsString[b]
		b = sec % 100 * 2
		tmp[1] = smallsString[b+1]
		tmp[0] = smallsString[b]
		// append to e.buf
		e.buf = append(e.buf, tmp[:]...)
	case TimeFormatUnixWithMs:
		sec, nsec, _ := now()
		// 1595759807.105
		var tmp [14]byte
		// milli seconds
		a := int64(nsec) / 1000000
		b := a % 100 * 2
		tmp[13] = smallsString[b+1]
		tmp[12] = smallsString[b]
		tmp[11] = byte('0' + a/100)
		tmp[10] = '.'
		// seconds
		b = sec % 100 * 2
		sec /= 100
		tmp[9] = smallsString[b+1]
		tmp[8] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[7] = smallsString[b+1]
		tmp[6] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[5] = smallsString[b+1]
		tmp[4] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[3] = smallsString[b+1]
		tmp[2] = smallsString[b]
		b = sec % 100 * 2
		tmp[1] = smallsString[b+1]
		tmp[0] = smallsString[b]
		// append to e.buf
		e.buf = append(e.buf, tmp[:]...)
	default:
		e.buf = append(e.buf, '"')
		e.buf = timeNow().AppendFormat(e.buf, l.TimeFormat)
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
	case TraceLevel:
		e.buf = append(e.buf, ",\"level\":\"trace\""...)
	case FatalLevel:
		e.buf = append(e.buf, ",\"level\":\"fatal\""...)
	case PanicLevel:
		e.buf = append(e.buf, ",\"level\":\"panic\""...)
	}
	// context
	if l.Context != nil {
		e.buf = append(e.buf, l.Context...)
	}
	return e
}

// Time append append t formated as string using time.RFC3339Nano.
func (e *Entry) Time(key string, t time.Time) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	e.buf = t.AppendFormat(e.buf, "2006-01-02T15:04:05.999Z07:00")
	e.buf = append(e.buf, '"')
	return e
}

// TimeFormat append append t formated as string using timefmt.
func (e *Entry) TimeFormat(key string, timefmt string, t time.Time) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	switch timefmt {
	case TimeFormatUnix:
		e.buf = strconv.AppendInt(e.buf, t.Unix(), 10)
	case TimeFormatUnixMs:
		e.buf = strconv.AppendInt(e.buf, t.UnixNano()/1000000, 10)
	case TimeFormatUnixWithMs:
		e.buf = strconv.AppendInt(e.buf, t.Unix(), 10)
		e.buf = append(e.buf, '.')
		e.buf = strconv.AppendInt(e.buf, t.UnixNano()/1000000%1000, 10)
	default:
		e.buf = append(e.buf, '"')
		e.buf = t.AppendFormat(e.buf, timefmt)
		e.buf = append(e.buf, '"')
	}
	return e
}

// Times append append a formated as string array using time.RFC3339Nano.
func (e *Entry) Times(key string, a []time.Time) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, t := range a {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = append(e.buf, '"')
		e.buf = t.AppendFormat(e.buf, time.RFC3339Nano)
		e.buf = append(e.buf, '"')
	}
	e.buf = append(e.buf, ']')

	return e
}

// TimesFormat append append a formated as string array using timefmt.
func (e *Entry) TimesFormat(key string, timefmt string, a []time.Time) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, t := range a {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		switch timefmt {
		case TimeFormatUnix:
			e.buf = strconv.AppendInt(e.buf, t.Unix(), 10)
		case TimeFormatUnixMs:
			e.buf = strconv.AppendInt(e.buf, t.UnixNano()/1000000, 10)
		case TimeFormatUnixWithMs:
			e.buf = strconv.AppendInt(e.buf, t.Unix(), 10)
			e.buf = append(e.buf, '.')
			e.buf = strconv.AppendInt(e.buf, t.UnixNano()/1000000%1000, 10)
		default:
			e.buf = append(e.buf, '"')
			e.buf = t.AppendFormat(e.buf, timefmt)
			e.buf = append(e.buf, '"')
		}
	}
	e.buf = append(e.buf, ']')

	return e
}

// Bool append append the val as a bool to the entry.
func (e *Entry) Bool(key string, b bool) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	e.buf = strconv.AppendBool(e.buf, b)
	return e
}

// Bools adds the field key with val as a []bool to the entry.
func (e *Entry) Bools(key string, b []bool) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, a := range b {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendBool(e.buf, a)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Dur adds the field key with duration d to the entry.
func (e *Entry) Dur(key string, d time.Duration) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	if d < 0 {
		d = -d
		e.buf = append(e.buf, '-')
	}
	e.buf = strconv.AppendInt(e.buf, int64(d/time.Millisecond), 10)
	if n := (d % time.Millisecond); n != 0 {
		var tmp [7]byte
		b := n % 100 * 2
		n /= 100
		tmp[6] = smallsString[b+1]
		tmp[5] = smallsString[b]
		b = n % 100 * 2
		n /= 100
		tmp[4] = smallsString[b+1]
		tmp[3] = smallsString[b]
		b = n % 100 * 2
		tmp[2] = smallsString[b+1]
		tmp[1] = smallsString[b]
		tmp[0] = '.'
		e.buf = append(e.buf, tmp[:]...)
	}
	return e
}

// TimeDiff adds the field key with positive duration between time t and start.
// If time t is not greater than start, duration will be 0.
// Duration format follows the same principle as Dur().
func (e *Entry) TimeDiff(key string, t time.Time, start time.Time) *Entry {
	if e == nil {
		return e
	}
	var d time.Duration
	if t.After(start) {
		d = t.Sub(start)
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	e.buf = strconv.AppendInt(e.buf, int64(d/time.Millisecond), 10)
	if n := (d % time.Millisecond); n != 0 {
		var tmp [7]byte
		b := n % 100 * 2
		n /= 100
		tmp[6] = smallsString[b+1]
		tmp[5] = smallsString[b]
		b = n % 100 * 2
		n /= 100
		tmp[4] = smallsString[b+1]
		tmp[3] = smallsString[b]
		b = n % 100 * 2
		tmp[2] = smallsString[b+1]
		tmp[1] = smallsString[b]
		tmp[0] = '.'
		e.buf = append(e.buf, tmp[:]...)
	}
	return e
}

// Durs adds the field key with val as a []time.Duration to the entry.
func (e *Entry) Durs(key string, d []time.Duration) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, a := range d {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		if a < 0 {
			a = -a
			e.buf = append(e.buf, '-')
		}
		e.buf = strconv.AppendInt(e.buf, int64(a/time.Millisecond), 10)
		if n := (a % time.Millisecond); n != 0 {
			var tmp [7]byte
			b := n % 100 * 2
			n /= 100
			tmp[6] = smallsString[b+1]
			tmp[5] = smallsString[b]
			b = n % 100 * 2
			n /= 100
			tmp[4] = smallsString[b+1]
			tmp[3] = smallsString[b]
			b = n % 100 * 2
			tmp[2] = smallsString[b+1]
			tmp[1] = smallsString[b]
			tmp[0] = '.'
			e.buf = append(e.buf, tmp[:]...)
		}
	}
	e.buf = append(e.buf, ']')
	return e
}

// Err adds the field "error" with serialized err to the entry.
func (e *Entry) Err(err error) *Entry {
	return e.AnErr("error", err)
}

// AnErr adds the field key with serialized err to the logger context.
func (e *Entry) AnErr(key string, err error) *Entry {
	if e == nil {
		return nil
	}

	if err == nil {
		e.buf = append(e.buf, ',', '"')
		e.buf = append(e.buf, key...)
		e.buf = append(e.buf, "\":null"...)
		return e
	}

	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	if o, ok := err.(ObjectMarshaler); ok {
		o.MarshalObject(e)
	} else {
		e.buf = append(e.buf, '"')
		e.string(err.Error())
		e.buf = append(e.buf, '"')
	}
	return e
}

// Errs adds the field key with errs as an array of serialized errors to the entry.
func (e *Entry) Errs(key string, errs []error) *Entry {
	if e == nil {
		return nil
	}

	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, err := range errs {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		if err == nil {
			e.buf = append(e.buf, "null"...)
		} else {
			e.buf = append(e.buf, '"')
			e.string(err.Error())
			e.buf = append(e.buf, '"')
		}
	}
	e.buf = append(e.buf, ']')
	return e
}

// Float64 adds the field key with f as a float64 to the entry.
func (e *Entry) Float64(key string, f float64) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	e.buf = strconv.AppendFloat(e.buf, f, 'f', -1, 64)
	return e
}

// Floats64 adds the field key with f as a []float64 to the entry.
func (e *Entry) Floats64(key string, f []float64) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, a := range f {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendFloat(e.buf, a, 'f', -1, 64)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Floats32 adds the field key with f as a []float32 to the entry.
func (e *Entry) Floats32(key string, f []float32) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, a := range f {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendFloat(e.buf, float64(a), 'f', -1, 32)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Int64 adds the field key with i as a int64 to the entry.
func (e *Entry) Int64(key string, i int64) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	e.buf = strconv.AppendInt(e.buf, i, 10)
	return e
}

// Uint adds the field key with i as a uint to the entry.
func (e *Entry) Uint(key string, i uint) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	e.buf = strconv.AppendUint(e.buf, uint64(i), 10)
	return e
}

// Uint64 adds the field key with i as a uint64 to the entry.
func (e *Entry) Uint64(key string, i uint64) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	e.buf = strconv.AppendUint(e.buf, i, 10)
	return e
}

// Float32 adds the field key with f as a float32 to the entry.
func (e *Entry) Float32(key string, f float32) *Entry {
	return e.Float64(key, float64(f))
}

// Int adds the field key with i as a int to the entry.
func (e *Entry) Int(key string, i int) *Entry {
	return e.Int64(key, int64(i))
}

// Int32 adds the field key with i as a int32 to the entry.
func (e *Entry) Int32(key string, i int32) *Entry {
	return e.Int64(key, int64(i))
}

// Int16 adds the field key with i as a int16 to the entry.
func (e *Entry) Int16(key string, i int16) *Entry {
	return e.Int64(key, int64(i))
}

// Int8 adds the field key with i as a int8 to the entry.
func (e *Entry) Int8(key string, i int8) *Entry {
	return e.Int64(key, int64(i))
}

// Uint32 adds the field key with i as a uint32 to the entry.
func (e *Entry) Uint32(key string, i uint32) *Entry {
	return e.Uint64(key, uint64(i))
}

// Uint16 adds the field key with i as a uint16 to the entry.
func (e *Entry) Uint16(key string, i uint16) *Entry {
	return e.Uint64(key, uint64(i))
}

// Uint8 adds the field key with i as a uint8 to the entry.
func (e *Entry) Uint8(key string, i uint8) *Entry {
	return e.Uint64(key, uint64(i))
}

// Ints64 adds the field key with i as a []int64 to the entry.
func (e *Entry) Ints64(key string, a []int64) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, n := range a {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendInt(e.buf, n, 10)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Ints32 adds the field key with i as a []int32 to the entry.
func (e *Entry) Ints32(key string, a []int32) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, n := range a {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendInt(e.buf, int64(n), 10)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Ints16 adds the field key with i as a []int16 to the entry.
func (e *Entry) Ints16(key string, a []int16) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, n := range a {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendInt(e.buf, int64(n), 10)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Ints8 adds the field key with i as a []int8 to the entry.
func (e *Entry) Ints8(key string, a []int8) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, n := range a {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendInt(e.buf, int64(n), 10)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Ints adds the field key with i as a []int to the entry.
func (e *Entry) Ints(key string, a []int) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, n := range a {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendInt(e.buf, int64(n), 10)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Uints64 adds the field key with i as a []uint64 to the entry.
func (e *Entry) Uints64(key string, a []uint64) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, n := range a {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendUint(e.buf, n, 10)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Uints32 adds the field key with i as a []uint32 to the entry.
func (e *Entry) Uints32(key string, a []uint32) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, n := range a {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendUint(e.buf, uint64(n), 10)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Uints16 adds the field key with i as a []uint16 to the entry.
func (e *Entry) Uints16(key string, a []uint16) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, n := range a {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendUint(e.buf, uint64(n), 10)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Uints8 adds the field key with i as a []uint8 to the entry.
func (e *Entry) Uints8(key string, a []uint8) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, n := range a {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendUint(e.buf, uint64(n), 10)
	}
	e.buf = append(e.buf, ']')
	return e
}

// Uints adds the field key with i as a []uint to the entry.
func (e *Entry) Uints(key string, a []uint) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, n := range a {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = strconv.AppendUint(e.buf, uint64(n), 10)
	}
	e.buf = append(e.buf, ']')
	return e
}

// RawJSON adds already encoded JSON to the log line under key.
func (e *Entry) RawJSON(key string, b []byte) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	e.buf = append(e.buf, b...)
	return e
}

// RawJSONStr adds already encoded JSON String to the log line under key.
func (e *Entry) RawJSONStr(key string, s string) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	e.buf = append(e.buf, s...)
	return e
}

// Str adds the field key with val as a string to the entry.
func (e *Entry) Str(key string, val string) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	e.string(val)
	e.buf = append(e.buf, '"')
	return e
}

// StrInt adds the field key with integer val as a string to the entry.
func (e *Entry) StrInt(key string, val int64) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	e.buf = strconv.AppendInt(e.buf, val, 10)
	e.buf = append(e.buf, '"')
	return e
}

// Stringer adds the field key with val.String() to the entry.
func (e *Entry) Stringer(key string, val fmt.Stringer) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	if val != nil {
		e.buf = append(e.buf, '"')
		e.string(val.String())
		e.buf = append(e.buf, '"')
	} else {
		e.buf = append(e.buf, "null"...)
	}
	return e
}

// GoStringer adds the field key with val.GoStringer() to the entry.
func (e *Entry) GoStringer(key string, val fmt.GoStringer) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	if val != nil {
		e.buf = append(e.buf, '"')
		e.string(val.GoString())
		e.buf = append(e.buf, '"')
	} else {
		e.buf = append(e.buf, "null"...)
	}
	return e
}

// Strs adds the field key with vals as a []string to the entry.
func (e *Entry) Strs(key string, vals []string) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '[')
	for i, val := range vals {
		if i != 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = append(e.buf, '"')
		e.string(val)
		e.buf = append(e.buf, '"')
	}
	e.buf = append(e.buf, ']')
	return e
}

// Byte adds the field key with val as a byte to the entry.
func (e *Entry) Byte(key string, val byte) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	switch val {
	case '"':
		e.buf = append(e.buf, "\"\\\"\""...)
	case '\\':
		e.buf = append(e.buf, "\"\\\\\""...)
	case '\n':
		e.buf = append(e.buf, "\"\\n\""...)
	case '\r':
		e.buf = append(e.buf, "\"\\r\""...)
	case '\t':
		e.buf = append(e.buf, "\"\\t\""...)
	case '\f':
		e.buf = append(e.buf, "\"\\u000c\""...)
	case '\b':
		e.buf = append(e.buf, "\"\\u0008\""...)
	case '<':
		e.buf = append(e.buf, "\"\\u003c\""...)
	case '\'':
		e.buf = append(e.buf, "\"\\u0027\""...)
	case 0:
		e.buf = append(e.buf, "\"\\u0000\""...)
	default:
		e.buf = append(e.buf, '"', val, '"')
	}
	return e
}

// Bytes adds the field key with val as a string to the entry.
func (e *Entry) Bytes(key string, val []byte) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	e.bytes(val)
	e.buf = append(e.buf, '"')
	return e
}

// BytesOrNil adds the field key with val as a string or nil to the entry.
func (e *Entry) BytesOrNil(key string, val []byte) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	if val == nil {
		e.buf = append(e.buf, "null"...)
	} else {
		e.buf = append(e.buf, '"')
		e.bytes(val)
		e.buf = append(e.buf, '"')
	}
	return e
}

const hex = "0123456789abcdef"

// Hex adds the field key with val as a hex string to the entry.
func (e *Entry) Hex(key string, val []byte) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	for _, v := range val {
		e.buf = append(e.buf, hex[v>>4], hex[v&0x0f])
	}
	e.buf = append(e.buf, '"')
	return e
}

// Xid adds the field key with xid.ID as a base32 string to the entry.
func (e *Entry) Xid(key string, xid [12]byte) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	e.buf = append(e.buf, (XID(xid)).String()...)
	e.buf = append(e.buf, '"')

	return e
}

// IPAddr adds IPv4 or IPv6 Address to the entry.
func (e *Entry) IPAddr(key string, ip net.IP) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	if ip4 := ip.To4(); ip4 != nil {
		e.buf = strconv.AppendInt(e.buf, int64(ip4[0]), 10)
		e.buf = append(e.buf, '.')
		e.buf = strconv.AppendInt(e.buf, int64(ip4[1]), 10)
		e.buf = append(e.buf, '.')
		e.buf = strconv.AppendInt(e.buf, int64(ip4[2]), 10)
		e.buf = append(e.buf, '.')
		e.buf = strconv.AppendInt(e.buf, int64(ip4[3]), 10)
	} else {
		e.buf = append(e.buf, ip.String()...)
	}
	e.buf = append(e.buf, '"')
	return e
}

// IPPrefix adds IPv4 or IPv6 Prefix (address and mask) to the entry.
func (e *Entry) IPPrefix(key string, pfx net.IPNet) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	e.buf = append(e.buf, pfx.String()...)
	e.buf = append(e.buf, '"')
	return e
}

// MACAddr adds MAC address to the entry.
func (e *Entry) MACAddr(key string, ha net.HardwareAddr) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	for i, c := range ha {
		if i > 0 {
			e.buf = append(e.buf, ':')
		}
		e.buf = append(e.buf, hex[c>>4])
		e.buf = append(e.buf, hex[c&0xF])
	}
	e.buf = append(e.buf, '"')
	return e
}

// Caller adds the file:line of the "caller" key.
// If depth is negative, adds the full /path/to/file:line of the "caller" key.
func (e *Entry) Caller(depth int) *Entry {
	if e != nil {
		var full bool
		var rpc [1]uintptr
		if depth < 0 {
			depth, full = -depth, true
		}
		e.caller(callers(depth, rpc[:]), rpc[:], full)
	}
	return e
}

// Stack enables stack trace printing for the error passed to Err().
func (e *Entry) Stack() *Entry {
	if e != nil {
		e.buf = append(e.buf, ",\"stack\":\""...)
		e.bytes(stacks(false))
		e.buf = append(e.buf, '"')
	}
	return e
}

// Enabled return false if the entry is going to be filtered out by log level.
func (e *Entry) Enabled() bool {
	return e != nil
}

// Discard disables the entry so Msg(f) won't print it.
func (e *Entry) Discard() *Entry {
	if e == nil {
		return e
	}
	if cap(e.buf) <= bbcap {
		epool.Put(e)
	}
	return nil
}

var notTest = true

// Msg sends the entry with msg added as the message field if not empty.
func (e *Entry) Msg(msg string) {
	if e == nil {
		return
	}
	if msg != "" {
		e.buf = append(e.buf, ",\"message\":\""...)
		e.string(msg)
		e.buf = append(e.buf, "\"}\n"...)
	} else {
		e.buf = append(e.buf, '}', '\n')
	}
	_, _ = e.w.WriteEntry(e)
	if (e.Level == FatalLevel) && notTest {
		os.Exit(255)
	}
	if (e.Level == PanicLevel) && notTest {
		panic(msg)
	}
	if cap(e.buf) <= bbcap {
		epool.Put(e)
	}
}

type bb struct {
	B []byte
}

func (b *bb) Write(p []byte) (int, error) {
	b.B = append(b.B, p...)
	return len(p), nil
}

var bbpool = sync.Pool{
	New: func() interface{} {
		return new(bb)
	},
}

// Msgf sends the entry with formatted msg added as the message field if not empty.
func (e *Entry) Msgf(format string, v ...interface{}) {
	if e == nil {
		return
	}
	b := bbpool.Get().(*bb)
	b.B = b.B[:0]
	e.buf = append(e.buf, ",\"message\":\""...)
	fmt.Fprintf(b, format, v...)
	e.bytes(b.B)
	e.buf = append(e.buf, '"')
	if cap(b.B) <= bbcap {
		bbpool.Put(b)
	}
	e.Msg("")
}

// Msgs sends the entry with msgs added as the message field if not empty.
func (e *Entry) Msgs(args ...interface{}) {
	if e == nil {
		return
	}
	b := bbpool.Get().(*bb)
	b.B = b.B[:0]
	e.buf = append(e.buf, ",\"message\":\""...)
	fmt.Fprint(b, args...)
	e.bytes(b.B)
	e.buf = append(e.buf, '"')
	if cap(b.B) <= bbcap {
		bbpool.Put(b)
	}
	e.Msg("")
}

func (e *Entry) caller(n int, rpc []uintptr, fullpath bool) {
	if n < 1 {
		return
	}
	frame, _ := runtime.CallersFrames(rpc).Next()
	file := frame.File
	if !fullpath {
		var i int
		for i = len(file) - 1; i >= 0; i-- {
			if file[i] == '/' {
				break
			}
		}
		if i > 0 {
			file = file[i+1:]
		}
	}

	e.buf = append(e.buf, ",\"caller\":\""...)
	e.buf = append(e.buf, file...)
	e.buf = append(e.buf, ':')
	e.buf = strconv.AppendInt(e.buf, int64(frame.Line), 10)
	e.buf = append(e.buf, "\",\"goid\":"...)
	e.buf = strconv.AppendInt(e.buf, int64(goid()), 10)
}

var escapes = [256]bool{
	'"':  true,
	'<':  true,
	'\'': true,
	'\\': true,
	'\b': true,
	'\f': true,
	'\n': true,
	'\r': true,
	'\t': true,
}

func (e *Entry) escapeb(b []byte) {
	n := len(b)
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
}

func (e *Entry) escapes(s string) {
	n := len(s)
	j := 0
	if n > 0 {
		// Hint the compiler to remove bounds checks in the loop below.
		_ = s[n-1]
	}
	for i := 0; i < n; i++ {
		switch s[i] {
		case '"':
			e.buf = append(e.buf, s[j:i]...)
			e.buf = append(e.buf, '\\', '"')
			j = i + 1
		case '\\':
			e.buf = append(e.buf, s[j:i]...)
			e.buf = append(e.buf, '\\', '\\')
			j = i + 1
		case '\n':
			e.buf = append(e.buf, s[j:i]...)
			e.buf = append(e.buf, '\\', 'n')
			j = i + 1
		case '\r':
			e.buf = append(e.buf, s[j:i]...)
			e.buf = append(e.buf, '\\', 'r')
			j = i + 1
		case '\t':
			e.buf = append(e.buf, s[j:i]...)
			e.buf = append(e.buf, '\\', 't')
			j = i + 1
		case '\f':
			e.buf = append(e.buf, s[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '0', 'c')
			j = i + 1
		case '\b':
			e.buf = append(e.buf, s[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '0', '8')
			j = i + 1
		case '<':
			e.buf = append(e.buf, s[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '3', 'c')
			j = i + 1
		case '\'':
			e.buf = append(e.buf, s[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '2', '7')
			j = i + 1
		case 0:
			e.buf = append(e.buf, s[j:i]...)
			e.buf = append(e.buf, '\\', 'u', '0', '0', '0', '0')
			j = i + 1
		}
	}
	e.buf = append(e.buf, s[j:]...)
}

func (e *Entry) string(s string) {
	for _, c := range []byte(s) {
		if escapes[c] {
			e.escapes(s)
			return
		}
	}
	e.buf = append(e.buf, s...)
}

func (e *Entry) bytes(b []byte) {
	for _, c := range b {
		if escapes[c] {
			e.escapeb(b)
			return
		}
	}
	e.buf = append(e.buf, b...)
}

// Interface adds the field key with i marshaled using reflection.
func (e *Entry) Interface(key string, i interface{}) *Entry {
	if e == nil {
		return nil
	}

	if o, ok := i.(ObjectMarshaler); ok {
		return e.Object(key, o)
	}

	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	b := bbpool.Get().(*bb)
	b.B = b.B[:0]
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	err := enc.Encode(i)
	if err != nil {
		b.B = b.B[:0]
		fmt.Fprintf(b, "marshaling error: %+v", err)
	} else {
		b.B = b.B[:len(b.B)-1]
	}
	e.bytes(b.B)
	e.buf = append(e.buf, '"')
	if cap(b.B) <= bbcap {
		bbpool.Put(b)
	}

	return e
}

// Object marshals an object that implement the ObjectMarshaler interface.
func (e *Entry) Object(key string, obj ObjectMarshaler) *Entry {
	if e == nil {
		return nil
	}

	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
	if obj == nil || (*[2]uintptr)(unsafe.Pointer(&obj))[1] == 0 {
		e.buf = append(e.buf, "null"...)
		return e
	}

	n := len(e.buf)
	obj.MarshalObject(e)
	if n < len(e.buf) {
		e.buf[n] = '{'
		e.buf = append(e.buf, '}')
	} else {
		e.buf = append(e.buf, "null"...)
	}

	return e
}

// Func allows an anonymous func to run only if the entry is enabled.
func (e *Entry) Func(f func(e *Entry)) *Entry {
	if e != nil {
		f(e)
	}
	return e
}

// EmbedObject marshals and Embeds an object that implement the ObjectMarshaler interface.
func (e *Entry) EmbedObject(obj ObjectMarshaler) *Entry {
	if e == nil {
		return nil
	}

	if obj != nil && (*[2]uintptr)(unsafe.Pointer(&obj))[1] != 0 {
		obj.MarshalObject(e)
	}
	return e
}

// KeysAndValues sends keysAndValues to Entry
func (e *Entry) KeysAndValues(keysAndValues ...interface{}) *Entry {
	if e == nil {
		return nil
	}
	var key string
	for i, v := range keysAndValues {
		if i%2 == 0 {
			key, _ = v.(string)
			continue
		}
		if v == nil {
			e.buf = append(e.buf, ',', '"')
			e.buf = append(e.buf, key...)
			e.buf = append(e.buf, '"', ':')
			e.buf = append(e.buf, "null"...)
			continue
		}
		switch v := v.(type) {
		case ObjectMarshaler:
			e.buf = append(e.buf, ',', '"')
			e.buf = append(e.buf, key...)
			e.buf = append(e.buf, '"', ':')
			v.MarshalObject(e)
		case Context:
			e.Dict(key, v)
		case []time.Duration:
			e.Durs(key, v)
		case time.Duration:
			e.Dur(key, v)
		case time.Time:
			e.Time(key, v)
		case net.HardwareAddr:
			e.MACAddr(key, v)
		case net.IP:
			e.IPAddr(key, v)
		case net.IPNet:
			e.IPPrefix(key, v)
		case json.RawMessage:
			e.buf = append(e.buf, ',', '"')
			e.buf = append(e.buf, key...)
			e.buf = append(e.buf, '"', ':')
			e.buf = append(e.buf, v...)
		case []bool:
			e.Bools(key, v)
		case []byte:
			e.Bytes(key, v)
		case []error:
			e.Errs(key, v)
		case []float32:
			e.Floats32(key, v)
		case []float64:
			e.Floats64(key, v)
		case []string:
			e.Strs(key, v)
		case string:
			e.Str(key, v)
		case bool:
			e.Bool(key, v)
		case error:
			e.AnErr(key, v)
		case float32:
			e.Float32(key, v)
		case float64:
			e.Float64(key, v)
		case int16:
			e.Int16(key, v)
		case int32:
			e.Int32(key, v)
		case int64:
			e.Int64(key, v)
		case int8:
			e.Int8(key, v)
		case int:
			e.Int(key, v)
		case uint16:
			e.Uint16(key, v)
		case uint32:
			e.Uint32(key, v)
		case uint64:
			e.Uint64(key, v)
		case uint8:
			e.Uint8(key, v)
		case fmt.GoStringer:
			e.GoStringer(key, v)
		case fmt.Stringer:
			e.Stringer(key, v)
		default:
			e.Interface(key, v)
		}
	}
	return e
}

// Fields type, used to pass to `Fields`.
type Fields map[string]interface{}

// Fields is a helper function to use a map to set fields using type assertion.
func (e *Entry) Fields(fields Fields) *Entry {
	if e == nil {
		return nil
	}
	for k, v := range fields {
		if v == nil {
			e.buf = append(e.buf, ',', '"')
			e.buf = append(e.buf, k...)
			e.buf = append(e.buf, "\":null"...)
			continue
		}
		switch v := v.(type) {
		case ObjectMarshaler:
			e.buf = append(e.buf, ',', '"')
			e.buf = append(e.buf, k...)
			e.buf = append(e.buf, '"', ':')
			v.MarshalObject(e)
		case Context:
			e.Dict(k, v)
		case []time.Duration:
			e.Durs(k, v)
		case time.Duration:
			e.Dur(k, v)
		case time.Time:
			e.Time(k, v)
		case net.HardwareAddr:
			e.MACAddr(k, v)
		case net.IP:
			e.IPAddr(k, v)
		case net.IPNet:
			e.IPPrefix(k, v)
		case json.RawMessage:
			e.buf = append(e.buf, ',', '"')
			e.buf = append(e.buf, k...)
			e.buf = append(e.buf, '"', ':')
			e.buf = append(e.buf, v...)
		case []bool:
			e.Bools(k, v)
		case []byte:
			e.Bytes(k, v)
		case []error:
			e.Errs(k, v)
		case []float32:
			e.Floats32(k, v)
		case []float64:
			e.Floats64(k, v)
		case []string:
			e.Strs(k, v)
		case string:
			e.Str(k, v)
		case bool:
			e.Bool(k, v)
		case error:
			e.AnErr(k, v)
		case float32:
			e.Float32(k, v)
		case float64:
			e.Float64(k, v)
		case int16:
			e.Int16(k, v)
		case int32:
			e.Int32(k, v)
		case int64:
			e.Int64(k, v)
		case int8:
			e.Int8(k, v)
		case int:
			e.Int(k, v)
		case uint16:
			e.Uint16(k, v)
		case uint32:
			e.Uint32(k, v)
		case uint64:
			e.Uint64(k, v)
		case uint8:
			e.Uint8(k, v)
		case fmt.GoStringer:
			e.GoStringer(k, v)
		case fmt.Stringer:
			e.Stringer(k, v)
		default:
			e.Interface(k, v)
		}
	}
	return e
}

// Context represents contextual fields.
type Context []byte

// NewContext starts a new contextual entry.
func NewContext(dst []byte) (e *Entry) {
	e = new(Entry)
	e.buf = dst
	return
}

// Value builds the contextual fields.
func (e *Entry) Value() Context {
	return e.buf
}

// Context sends the contextual fields to entry.
func (e *Entry) Context(ctx Context) *Entry {
	if e == nil {
		return nil
	}
	if len(ctx) != 0 {
		e.buf = append(e.buf, ctx...)
	}
	return e
}

// Dict sends the contextual fields with key to entry.
func (e *Entry) Dict(key string, ctx Context) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '{')
	if len(ctx) > 0 {
		e.buf = append(e.buf, ctx[1:]...)
	}
	e.buf = append(e.buf, '}')
	return e
}

// stacks is a wrapper for runtime.Stack that attempts to recover the data for all goroutines.
func stacks(all bool) (trace []byte) {
	// We don't know how big the traces are, so grow a few times if they don't fit. Start large, though.
	n := 10000
	if all {
		n = 100000
	}
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, all)
		if nbytes < len(trace) {
			trace = trace[:nbytes]
			break
		}
		n *= 2
	}
	return
}

// wlprintf is a helper function for tests
func wlprintf(w Writer, level Level, format string, args ...interface{}) (int, error) {
	return w.WriteEntry(&Entry{
		Level: level,
		buf:   []byte(fmt.Sprintf(format, args...)),
	})
}

func b2s(b []byte) string { return *(*string)(unsafe.Pointer(&b)) }

//go:noescape
//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64)

//go:noescape
//go:linkname absDate time.absDate
func absDate(abs uint64, full bool) (year int, month time.Month, day int, yday int)

//go:noescape
//go:linkname absClock time.absClock
func absClock(abs uint64) (hour, min, sec int)

//go:noescape
//go:linkname callers runtime.callers
func callers(skip int, pcbuf []uintptr) int

// Fastrandn returns a pseudorandom uint32 in [0,n).
//
//go:noescape
//go:linkname Fastrandn runtime.fastrandn
func Fastrandn(x uint32) uint32
