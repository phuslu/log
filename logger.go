package log

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
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
	Writer:     os.Stderr,
}

// A Logger represents an active logging object that generates lines of JSON output to an io.Writer.
type Logger struct {
	// Level defines log levels.
	Level Level

	// Caller determines if adds the file:line of the "caller" key.
	Caller int

	// TimeField defines the time filed name in output.  It uses "time" in if empty.
	TimeField string

	// TimeFormat specifies the time format in output. It uses time.RFC3389 in if empty.
	// If set with `TimeFormatUnix`, `TimeFormatUnixMs`, times are formated as UNIX timestamp.
	TimeFormat string

	// Writer specifies the writer of output. It uses os.Stderr in if empty.
	Writer io.Writer
}

const (
	// TimeFormatUnix defines a time format that makes time fields to be
	// serialized as Unix timestamp integers.
	TimeFormatUnix = "\x01"

	// TimeFormatUnixMs defines a time format that makes time fields to be
	// serialized as Unix timestamp integers in milliseconds.
	TimeFormatUnixMs = "\x02"
)

// Event represents a log event. It is instanced by one of the level method of Logger and finalized by the Msg or Msgf method.
type Event struct {
	buf      []byte
	w        io.Writer
	stack    bool
	stackall bool
	exit     bool
	panic    bool
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

// Panic starts a new message with panic level.
func Panic() (e *Event) {
	e = DefaultLogger.header(PanicLevel)
	if e != nil && DefaultLogger.Caller > 0 {
		e.caller(runtime.Caller(DefaultLogger.Caller))
	}
	return
}

// Printf sends a log event without extra field. Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	e := DefaultLogger.header(noLevel)
	if e != nil && DefaultLogger.Caller > 0 {
		e.caller(runtime.Caller(DefaultLogger.Caller))
	}
	e.Msgf(format, v...)
}

// Debug starts a new message with debug level.
func (l *Logger) Debug() (e *Event) {
	e = l.header(DebugLevel)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	return
}

// Info starts a new message with info level.
func (l *Logger) Info() (e *Event) {
	e = l.header(InfoLevel)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	return
}

// Warn starts a new message with warning level.
func (l *Logger) Warn() (e *Event) {
	e = l.header(WarnLevel)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	return
}

// Error starts a new message with error level.
func (l *Logger) Error() (e *Event) {
	e = l.header(ErrorLevel)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	return
}

// Fatal starts a new message with fatal level.
func (l *Logger) Fatal() (e *Event) {
	e = l.header(FatalLevel)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	return
}

// Panic starts a new message with panic level.
func (l *Logger) Panic() (e *Event) {
	e = l.header(PanicLevel)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	return
}

// WithLevel starts a new message with level.
func (l *Logger) WithLevel(level Level) (e *Event) {
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

// Printf sends a log event without extra field. Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Printf(format string, v ...interface{}) {
	e := l.header(noLevel)
	if e != nil && l.Caller > 0 {
		e.caller(runtime.Caller(l.Caller))
	}
	e.Msgf(format, v...)
}

var epool = sync.Pool{
	New: func() interface{} {
		return &Event{
			buf: make([]byte, 0, 500),
		}
	},
}

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

func (l *Logger) header(level Level) *Event {
	if uint32(level) < atomic.LoadUint32((*uint32)(&l.Level)) {
		return nil
	}
	e := epool.Get().(*Event)
	e.buf = e.buf[:0]

	switch level {
	default:
		e.stack = false
		e.stackall = false
		e.exit = false
		e.panic = false
	case FatalLevel:
		e.stack = true
		e.stackall = true
		e.exit = true
		e.panic = false
	case PanicLevel:
		e.stack = true
		e.stackall = false
		e.exit = false
		e.panic = true
	}
	if l.Writer != nil {
		e.w = l.Writer
	} else {
		e.w = os.Stderr
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
		n := len(e.buf)
		if timeOffset == 0 {
			// "2006-01-02T15:04:05.999Z"
			e.buf = e.buf[:n+26]
			e.buf[n+25] = '"'
			e.buf[n+24] = 'Z'
		} else {
			// "2006-01-02T15:04:05.999Z07:00"
			e.buf = e.buf[:n+31]
			e.buf[n+30] = '"'
			e.buf[n+29] = timeZone[5]
			e.buf[n+28] = timeZone[4]
			e.buf[n+27] = timeZone[3]
			e.buf[n+26] = timeZone[2]
			e.buf[n+25] = timeZone[1]
			e.buf[n+24] = timeZone[0]
		}
		sec, nsec := walltime()
		// date time
		sec += 9223372028715321600 + timeOffset // unixToInternal + internalToAbsolute + timeOffset
		year, month, day, _ := absDate(uint64(sec), true)
		hour, minute, second := absClock(uint64(sec))
		// year
		a := year / 100 * 2
		b := year % 100 * 2
		e.buf[n] = '"'
		e.buf[n+1] = smallsString[a]
		e.buf[n+2] = smallsString[a+1]
		e.buf[n+3] = smallsString[b]
		e.buf[n+4] = smallsString[b+1]
		// month
		month *= 2
		e.buf[n+5] = '-'
		e.buf[n+6] = smallsString[month]
		e.buf[n+7] = smallsString[month+1]
		// day
		day *= 2
		e.buf[n+8] = '-'
		e.buf[n+9] = smallsString[day]
		e.buf[n+10] = smallsString[day+1]
		// hour
		hour *= 2
		e.buf[n+11] = 'T'
		e.buf[n+12] = smallsString[hour]
		e.buf[n+13] = smallsString[hour+1]
		// minute
		minute *= 2
		e.buf[n+14] = ':'
		e.buf[n+15] = smallsString[minute]
		e.buf[n+16] = smallsString[minute+1]
		// second
		second *= 2
		e.buf[n+17] = ':'
		e.buf[n+18] = smallsString[second]
		e.buf[n+19] = smallsString[second+1]
		// milli seconds
		a = int(nsec) / 1000000
		b = a % 100 * 2
		e.buf[n+20] = '.'
		e.buf[n+21] = byte('0' + a/100)
		e.buf[n+22] = smallsString[b]
		e.buf[n+23] = smallsString[b+1]
	case TimeFormatUnix:
		// 1595759807
		n := len(e.buf)
		e.buf = e.buf[:n+10]
		sec, _ := walltime()
		// seconds
		b := sec % 100 * 2
		sec /= 100
		e.buf[n+9] = smallsString[b+1]
		e.buf[n+8] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		e.buf[n+7] = smallsString[b+1]
		e.buf[n+6] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		e.buf[n+5] = smallsString[b+1]
		e.buf[n+4] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		e.buf[n+3] = smallsString[b+1]
		e.buf[n+2] = smallsString[b]
		b = sec % 100 * 2
		e.buf[n+1] = smallsString[b+1]
		e.buf[n] = smallsString[b]
	case TimeFormatUnixMs:
		// 1595759807105
		n := len(e.buf)
		e.buf = e.buf[:n+13]
		sec, nsec := walltime()
		// seconds
		b := sec % 100 * 2
		sec /= 100
		e.buf[n+9] = smallsString[b+1]
		e.buf[n+8] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		e.buf[n+7] = smallsString[b+1]
		e.buf[n+6] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		e.buf[n+5] = smallsString[b+1]
		e.buf[n+4] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		e.buf[n+3] = smallsString[b+1]
		e.buf[n+2] = smallsString[b]
		b = sec % 100 * 2
		e.buf[n+1] = smallsString[b+1]
		e.buf[n] = smallsString[b]
		// milli seconds
		a := int64(nsec) / 1000000
		b = a % 100 * 2
		e.buf[n+10] = byte('0' + a/100)
		e.buf[n+11] = smallsString[b]
		e.buf[n+12] = smallsString[b+1]
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
	case FatalLevel:
		e.buf = append(e.buf, ",\"level\":\"fatal\""...)
	case PanicLevel:
		e.buf = append(e.buf, ",\"level\":\"panic\""...)
	}
	return e
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

// AnErr adds the field key with serialized err to the logger context.
func (e *Event) AnErr(key string, err error) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	if err == nil {
		e.buf = append(e.buf, "null"...)
	} else {
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

// RawJSONStr adds already encoded JSON String to the log line under key.
func (e *Event) RawJSONStr(key string, s string) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = append(e.buf, s...)
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

// Stringer adds the field key with val.String() to the event.
func (e *Event) Stringer(key string, val fmt.Stringer) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	if val != nil {
		e.string(val.String())
	} else {
		e.buf = append(e.buf, "null"...)
	}
	return e
}

// GoStringer adds the field key with val.GoStringer() to the event.
func (e *Event) GoStringer(key string, val fmt.GoStringer) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	if val != nil {
		e.string(val.GoString())
	} else {
		e.buf = append(e.buf, "null"...)
	}
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

// BytesOrNil adds the field key with val as a string or nil to the event.
func (e *Event) BytesOrNil(key string, val []byte) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	if val == nil {
		e.buf = append(e.buf, "null"...)
	} else {
		e.bytes(val)
	}
	return e
}

const hex = "0123456789abcdef"

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

const base32 = "0123456789abcdefghijklmnopqrstuv"

// Xid adds the field key with xid.ID as a base32 string to the event.
func (e *Event) Xid(key string, xid [12]byte) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	_ = xid[11]
	e.buf = append(e.buf,
		'"',
		base32[xid[0]>>3],
		base32[(xid[1]>>6)&0x1F|(xid[0]<<2)&0x1F],
		base32[(xid[1]>>1)&0x1F],
		base32[(xid[2]>>4)&0x1F|(xid[1]<<4)&0x1F],
		base32[xid[3]>>7|(xid[2]<<1)&0x1F],
		base32[(xid[3]>>2)&0x1F],
		base32[xid[4]>>5|(xid[3]<<3)&0x1F],
		base32[xid[4]&0x1F],
		base32[xid[5]>>3],
		base32[(xid[6]>>6)&0x1F|(xid[5]<<2)&0x1F],
		base32[(xid[6]>>1)&0x1F],
		base32[(xid[7]>>4)&0x1F|(xid[6]<<4)&0x1F],
		base32[xid[8]>>7|(xid[7]<<1)&0x1F],
		base32[(xid[8]>>2)&0x1F],
		base32[(xid[9]>>5)|(xid[8]<<3)&0x1F],
		base32[xid[9]&0x1F],
		base32[xid[10]>>3],
		base32[(xid[11]>>6)&0x1F|(xid[10]<<2)&0x1F],
		base32[(xid[11]>>1)&0x1F],
		base32[(xid[11]<<4)&0x1F],
		'"',
	)

	return e
}

// IPAddr adds IPv4 or IPv6 Address to the event
func (e *Event) IPAddr(key string, ip net.IP) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = append(e.buf, '"')
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

// IPPrefix adds IPv4 or IPv6 Prefix (address and mask) to the event
func (e *Event) IPPrefix(key string, pfx net.IPNet) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = append(e.buf, '"')
	e.buf = append(e.buf, pfx.String()...)
	e.buf = append(e.buf, '"')
	return e
}

// MACAddr adds MAC address to the event
func (e *Event) MACAddr(key string, ha net.HardwareAddr) *Event {
	if e == nil {
		return nil
	}
	e.key(key)
	e.buf = append(e.buf, '"')
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

// TimeDiff adds the field key with positive duration between time t and start.
// If time t is not greater than start, duration will be 0.
// Duration format follows the same principle as Dur().
func (e *Event) TimeDiff(key string, t time.Time, start time.Time) *Event {
	if e == nil {
		return e
	}
	var d time.Duration
	if t.After(start) {
		d = t.Sub(start)
	}
	e.key(key)
	e.buf = append(e.buf, '"')
	e.buf = append(e.buf, d.String()...)
	e.buf = append(e.buf, '"')
	return e
}

// Caller adds the file:line of the "caller" key.
func (e *Event) Caller(depth int) *Event {
	if e == nil {
		return nil
	}
	e.caller(runtime.Caller(depth))
	return e
}

// Stack enables stack trace printing for the error passed to Err().
func (e *Event) Stack(all bool) *Event {
	if e == nil {
		return nil
	}
	e.stack = true
	e.stackall = all
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
	if cap(e.buf) <= bbcap {
		epool.Put(e)
	}
	return nil
}

var notTest = true

// Msg sends the event with msg added as the message field if not empty.
func (e *Event) Msg(msg string) {
	if e == nil {
		return
	}
	if e.stack {
		e.buf = append(e.buf, ",\"stack\":"...)
		e.bytes(stacks(e.stackall))
	}
	if msg != "" {
		e.buf = append(e.buf, ",\"message\":"...)
		e.string(msg)
	}
	e.buf = append(e.buf, '}', '\n')
	e.w.Write(e.buf)
	if e.exit && notTest {
		os.Exit(255)
	}
	if e.panic && notTest {
		panic(msg)
	}
	if cap(e.buf) <= bbcap {
		epool.Put(e)
	}
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
	e.buf = append(e.buf, "\",\"goid\":"...)
	e.buf = strconv.AppendInt(e.buf, goid(), 10)
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

func (e *Event) escape(b []byte) {
	e.buf = append(e.buf, '"')
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
	e.buf = append(e.buf, '"')
}

func (e *Event) string(s string) {
	for _, c := range []byte(s) {
		if escapes[c] {
			sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
			b := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
				Data: sh.Data, Len: sh.Len, Cap: sh.Len,
			}))
			e.escape(b)
			return
		}
	}

	e.buf = append(e.buf, '"')
	e.buf = append(e.buf, s...)
	e.buf = append(e.buf, '"')

	return
}

func (e *Event) bytes(b []byte) {
	for _, c := range b {
		if escapes[c] {
			e.escape(b)
			return
		}
	}

	e.buf = append(e.buf, '"')
	e.buf = append(e.buf, b...)
	e.buf = append(e.buf, '"')

	return
}

type bb struct {
	B []byte
}

func (b *bb) Write(p []byte) (int, error) {
	b.B = append(b.B, p...)
	return len(p), nil
}

func (b *bb) Reset() {
	b.B = b.B[:0]
}

var bbpool = sync.Pool{
	New: func() interface{} {
		return new(bb)
	},
}

const bbcap = 1 << 16

// Interface adds the field key with i marshaled using reflection.
func (e *Event) Interface(key string, i interface{}) *Event {
	if e == nil {
		return nil
	}
	e.key(key)

	b := bbpool.Get().(*bb)
	b.Reset()

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	err := enc.Encode(i)
	if err != nil {
		e.string("marshaling error: " + err.Error())
	} else {
		e.bytes(b.B[:len(b.B)-1])
	}

	if cap(b.B) <= bbcap {
		bbpool.Put(b)
	}

	return e
}

// Msgf sends the event with formatted msg added as the message field if not empty.
func (e *Event) Msgf(format string, v ...interface{}) {
	if e == nil {
		return
	}

	b := bbpool.Get().(*bb)
	b.Reset()

	fmt.Fprintf(b, format, v...)
	e.Msg(*(*string)(unsafe.Pointer(&b.B)))

	if cap(b.B) <= bbcap {
		bbpool.Put(b)
	}
}

// Context represents contextual fields.
type Context []byte

// NewContext starts a new contextual event.
func NewContext() (e *Event) {
	e = epool.Get().(*Event)
	e.buf = e.buf[:0]
	return
}

// Value builds the contextual fields.
func (e *Event) Value() Context {
	return e.buf
}

// Context sends the contextual fields to event.
func (e *Event) Context(ctx Context) *Event {
	if e == nil {
		return nil
	}
	if len(ctx) != 0 {
		e.buf = append(e.buf, ctx...)
	}
	return e
}

// Dict sends the contextual fields with key to event.
func (e *Event) Dict(key string, ctx Context) *Event {
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

//go:noescape
//go:linkname absDate time.absDate
func absDate(abs uint64, full bool) (year int, month time.Month, day int, yday int)

//go:noescape
//go:linkname absClock time.absClock
func absClock(abs uint64) (hour, min, sec int)

// Fastrandn returns a pseudorandom uint32 in [0,n).
//
//go:noescape
//go:linkname Fastrandn runtime.fastrandn
func Fastrandn(x uint32) uint32

// Goid returns the current goroutine id.
// It exactly matches goroutine id of the stack trace.
func Goid() int64 { return goid() }
