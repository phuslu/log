package log

import (
	"io"
	"net"
	"os"
	"strconv"
	"sync"
)

// TSVLogger represents an active logging object that generates lines of TSV output to an io.Writer.
type TSVLogger struct {
	Separator byte
	Writer    io.Writer
}

// TSVEvent represents a tsv log event. It is instanced by one of TSVLogger and finalized by the Msg method.
type TSVEvent struct {
	buf []byte
	w   io.Writer
	sep byte
}

var tepool = sync.Pool{
	New: func() interface{} {
		return new(TSVEvent)
	},
}

// New starts a new tsv message.
func (l *TSVLogger) New() (e *TSVEvent) {
	e = tepool.Get().(*TSVEvent)
	e.sep = l.Separator
	if l.Writer != nil {
		e.w = l.Writer
	} else {
		e.w = os.Stderr
	}
	if e.sep == 0 {
		e.sep = '\t'
	}
	e.buf = e.buf[:0]
	return
}

// Timestamp adds the current time as UNIX timestamp
func (e *TSVEvent) Timestamp() *TSVEvent {
	i := len(e.buf)
	e.buf = append(e.buf, "0123456789\t"...)
	sec, _ := walltime()
	// separator
	e.buf[i+10] = e.sep
	// seconds
	is := sec % 100 * 2
	sec /= 100
	e.buf[i+9] = smallsString[is+1]
	e.buf[i+8] = smallsString[is]
	is = sec % 100 * 2
	sec /= 100
	e.buf[i+7] = smallsString[is+1]
	e.buf[i+6] = smallsString[is]
	is = sec % 100 * 2
	sec /= 100
	e.buf[i+5] = smallsString[is+1]
	e.buf[i+4] = smallsString[is]
	is = sec % 100 * 2
	sec /= 100
	e.buf[i+3] = smallsString[is+1]
	e.buf[i+2] = smallsString[is]
	is = sec % 100 * 2
	e.buf[i+1] = smallsString[is+1]
	e.buf[i] = smallsString[is]
	return e
}

// TimestampMS adds the current time with milliseconds as UNIX timestamp
func (e *TSVEvent) TimestampMS() *TSVEvent {
	i := len(e.buf)
	e.buf = append(e.buf, "0123456789000\t"...)
	sec, nsec := walltime()
	// separator
	e.buf[i+13] = e.sep
	// milli seconds
	a := int64(nsec) / 1000000
	is := a % 100 * 2
	e.buf[i+12] = smallsString[is+1]
	e.buf[i+11] = smallsString[is]
	e.buf[i+10] = byte('0' + a/100)
	// seconds
	is = sec % 100 * 2
	sec /= 100
	e.buf[i+9] = smallsString[is+1]
	e.buf[i+8] = smallsString[is]
	is = sec % 100 * 2
	sec /= 100
	e.buf[i+7] = smallsString[is+1]
	e.buf[i+6] = smallsString[is]
	is = sec % 100 * 2
	sec /= 100
	e.buf[i+5] = smallsString[is+1]
	e.buf[i+4] = smallsString[is]
	is = sec % 100 * 2
	sec /= 100
	e.buf[i+3] = smallsString[is+1]
	e.buf[i+2] = smallsString[is]
	is = sec % 100 * 2
	e.buf[i+1] = smallsString[is+1]
	e.buf[i] = smallsString[is]
	return e
}

// Bool append append the val as a bool to the event.
func (e *TSVEvent) Bool(b bool) *TSVEvent {
	if b {
		e.buf = append(e.buf, '1', e.sep)
	} else {
		e.buf = append(e.buf, '0', e.sep)
	}
	return e
}

// Float64 adds a float64 to the event.
func (e *TSVEvent) Float64(f float64) *TSVEvent {
	e.buf = strconv.AppendFloat(e.buf, f, 'f', -1, 64)
	e.buf = append(e.buf, e.sep)
	return e
}

// Int64 adds a int64 to the event.
func (e *TSVEvent) Int64(i int64) *TSVEvent {
	e.buf = strconv.AppendInt(e.buf, i, 10)
	e.buf = append(e.buf, e.sep)
	return e
}

// Uint64 adds a uint64 to the event.
func (e *TSVEvent) Uint64(i uint64) *TSVEvent {
	e.buf = strconv.AppendUint(e.buf, i, 10)
	e.buf = append(e.buf, e.sep)
	return e
}

// Float32 adds a float32 to the event.
func (e *TSVEvent) Float32(f float32) *TSVEvent {
	return e.Float64(float64(f))
}

// Int adds a int to the event.
func (e *TSVEvent) Int(i int) *TSVEvent {
	return e.Int64(int64(i))
}

// Int32 adds a int32 to the event.
func (e *TSVEvent) Int32(i int32) *TSVEvent {
	return e.Int64(int64(i))
}

// Int16 adds a int16 to the event.
func (e *TSVEvent) Int16(i int16) *TSVEvent {
	return e.Int64(int64(i))
}

// Int8 adds a int8 to the event.
func (e *TSVEvent) Int8(i int8) *TSVEvent {
	return e.Int64(int64(i))
}

// Uint32 adds a uint32 to the event.
func (e *TSVEvent) Uint32(i uint32) *TSVEvent {
	return e.Uint64(uint64(i))
}

// Uint16 adds a uint16 to the event.
func (e *TSVEvent) Uint16(i uint16) *TSVEvent {
	return e.Uint64(uint64(i))
}

// Uint8 adds a uint8 to the event.
func (e *TSVEvent) Uint8(i uint8) *TSVEvent {
	return e.Uint64(uint64(i))
}

// Str adds a string to the event.
func (e *TSVEvent) Str(val string) *TSVEvent {
	e.buf = append(e.buf, val...)
	e.buf = append(e.buf, e.sep)
	return e
}

// Bytes adds a bytes as string to the event.
func (e *TSVEvent) Bytes(val []byte) *TSVEvent {
	e.buf = append(e.buf, val...)
	e.buf = append(e.buf, e.sep)
	return e
}

// IPAddr adds IPv4 or IPv6 Address to the event
func (e *TSVEvent) IPAddr(ip net.IP) *TSVEvent {
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
	e.buf = append(e.buf, e.sep)
	return e
}

// Msg sends the event.
func (e *TSVEvent) Msg() {
	if len(e.buf) != 0 {
		e.buf[len(e.buf)-1] = '\n'
	}
	e.w.Write(e.buf)
	// see https://golang.org/issue/23199
	if cap(e.buf) <= 1<<16 {
		tepool.Put(e)
	}
}
