package log

import (
	"io"
	"strconv"
	"strings"
	"sync"
)

type TSVLogger struct {
	Separator byte
	Escape    bool
	Writer    io.Writer
}

type TSVEvent struct {
	logger TSVLogger
	sep    byte
	buf    []byte
}

var tepool = sync.Pool{
	New: func() interface{} {
		return new(TSVEvent)
	},
}

func (l TSVLogger) New() (e *TSVEvent) {
	e = tepool.Get().(*TSVEvent)
	e.logger = l
	if l.Separator != 0 {
		e.sep = l.Separator
	} else {
		e.sep = '\t'
	}
	e.buf = e.buf[:0]
	return
}

func (e *TSVEvent) Timestamp() *TSVEvent {
	if e == nil {
		return nil
	}
	e.buf = strconv.AppendInt(e.buf, timeNow().Unix(), 10)
	e.buf = append(e.buf, e.sep)
	return e
}

func (e *TSVEvent) Bool(b bool) *TSVEvent {
	if e == nil {
		return nil
	}
	if b {
		e.buf = append(e.buf, '1', e.sep)
	} else {
		e.buf = append(e.buf, '0', e.sep)
	}
	return e
}

func (e *TSVEvent) Float64(f float64) *TSVEvent {
	if e == nil {
		return nil
	}
	e.buf = strconv.AppendFloat(e.buf, f, 'f', -1, 64)
	e.buf = append(e.buf, e.sep)
	return e
}

func (e *TSVEvent) Int64(i int64) *TSVEvent {
	if e == nil {
		return nil
	}
	e.buf = strconv.AppendInt(e.buf, i, 10)
	e.buf = append(e.buf, e.sep)
	return e
}

func (e *TSVEvent) Uint64(i uint64) *TSVEvent {
	if e == nil {
		return nil
	}
	e.buf = strconv.AppendUint(e.buf, i, 10)
	e.buf = append(e.buf, e.sep)
	return e
}

func (e *TSVEvent) Float32(f float32) *TSVEvent {
	return e.Float64(float64(f))
}

func (e *TSVEvent) Int(i int) *TSVEvent {
	return e.Int64(int64(i))
}

func (e *TSVEvent) Int32(i int32) *TSVEvent {
	return e.Int64(int64(i))
}

func (e *TSVEvent) Int16(i int16) *TSVEvent {
	return e.Int64(int64(i))
}

func (e *TSVEvent) Int8(i int8) *TSVEvent {
	return e.Int64(int64(i))
}

func (e *TSVEvent) Uint32(i uint32) *TSVEvent {
	return e.Uint64(uint64(i))
}

func (e *TSVEvent) Uint16(i uint16) *TSVEvent {
	return e.Uint64(uint64(i))
}

func (e *TSVEvent) Uint8(i uint8) *TSVEvent {
	return e.Uint64(uint64(i))
}

func (e *TSVEvent) Str(val string) *TSVEvent {
	if e == nil {
		return nil
	}
	if e.logger.Escape && strings.IndexByte(val, e.logger.Separator) >= 0 {
		e.buf = append(e.buf, '"')
		e.buf = append(e.buf, val...)
		e.buf = append(e.buf, '"', e.sep)
	} else {
		e.buf = append(e.buf, val...)
		e.buf = append(e.buf, e.sep)
	}
	return e
}

func (e *TSVEvent) Bytes(val []byte) *TSVEvent {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, val...)
	e.buf = append(e.buf, e.sep)
	return e
}

func (e *TSVEvent) Send() {
	if e == nil || len(e.buf) == 0 {
		return
	}
	e.buf[len(e.buf)-1] = '\n'
	e.logger.Writer.Write(e.buf)
	tepool.Put(e)
}
