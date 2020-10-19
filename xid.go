package log

import (
	"sync/atomic"
	"time"
)

var counter = Fastrandn(4294967295)

type XID [12]byte

func NewXID() (x XID) {
	// timestamp
	sec, _ := walltime()
	x[0] = byte(sec >> 24)
	x[1] = byte(sec >> 16)
	x[2] = byte(sec >> 8)
	x[3] = byte(sec)
	// machine
	x[4] = machine[0]
	x[5] = machine[1]
	x[6] = machine[2]
	// pid
	x[7] = byte(pid >> 8)
	x[8] = byte(pid)
	// counter
	i := atomic.AddUint32(&counter, 1)
	x[9] = byte(i >> 16)
	x[10] = byte(i >> 8)
	x[11] = byte(i)
	return
}

func (x XID) Time() time.Time {
	return time.Unix(int64(x[0])<<32|int64(x[1])<<16|int64(x[2])<<8|int64(x[3]), 0)
}

// Pid returns the process id part of the id.
func (x XID) Pid() uint16 {
	return uint16(x[7])<<8 | uint16(x[8])
}

// Counter returns the incrementing value part of the id.
func (x XID) Counter() uint32 {
	return uint32(x[9])<<16 | uint32(x[10])<<8 | uint32(x[11])
}
