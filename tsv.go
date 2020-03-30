package log

import (
	"io"
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
	buf   []byte
	write func(p []byte) (n int, err error)
	sep   byte
}

var tepool = sync.Pool{
	New: func() interface{} {
		return new(TSVEvent)
	},
}

// New starts a new tsv message.
func (l TSVLogger) New() (e *TSVEvent) {
	e = tepool.Get().(*TSVEvent)
	e.sep = l.Separator
	if l.Writer != nil {
		e.write = l.Writer.Write
	} else {
		e.write = os.Stderr.Write
	}
	if e.sep == 0 {
		e.sep = '\t'
	}
	e.buf = e.buf[:0]
	return
}

// Timestamp adds the current time as UNIX timestamp
func (e *TSVEvent) Timestamp() *TSVEvent {
	sec, _ := walltime()
	e.buf = strconv.AppendInt(e.buf, sec, 10)
	e.buf = append(e.buf, e.sep)
	return e
}

// TimestampMS adds the current time with milliseconds as UNIX timestamp
func (e *TSVEvent) TimestampMS() *TSVEvent {
	sec, nsec := walltime()
	ms := int64(nsec / 1000000)
	e.buf = strconv.AppendInt(e.buf, sec, 10)
	switch {
	case ms < 10:
		e.buf = append(e.buf, '0', '0')
		e.buf = strconv.AppendInt(e.buf, ms, 10)
	case ms < 100:
		e.buf = append(e.buf, '0')
		e.buf = strconv.AppendInt(e.buf, ms, 10)
	default:
		e.buf = strconv.AppendInt(e.buf, ms, 10)
	}
	e.buf = append(e.buf, e.sep)
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

// Msg sends the event.
func (e *TSVEvent) Msg() {
	if e == nil {
		return
	}
	if len(e.buf) != 0 {
		e.buf[len(e.buf)-1] = '\n'
	}
	e.write(e.buf)
	tepool.Put(e)
}
