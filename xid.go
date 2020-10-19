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

const base32 = "0123456789abcdefghijklmnopqrstuv"

// Encode encodes the id using base32 encoding, writing 20 bytes to dst and return it.
func (x XID) Encode(dst []byte) []byte {
	dst = append(dst,
		base32[x[0]>>3],
		base32[(x[1]>>6)&0x1F|(x[0]<<2)&0x1F],
		base32[(x[1]>>1)&0x1F],
		base32[(x[2]>>4)&0x1F|(x[1]<<4)&0x1F],
		base32[x[3]>>7|(x[2]<<1)&0x1F],
		base32[(x[3]>>2)&0x1F],
		base32[x[4]>>5|(x[3]<<3)&0x1F],
		base32[x[4]&0x1F],
		base32[x[5]>>3],
		base32[(x[6]>>6)&0x1F|(x[5]<<2)&0x1F],
		base32[(x[6]>>1)&0x1F],
		base32[(x[7]>>4)&0x1F|(x[6]<<4)&0x1F],
		base32[x[8]>>7|(x[7]<<1)&0x1F],
		base32[(x[8]>>2)&0x1F],
		base32[(x[9]>>5)|(x[8]<<3)&0x1F],
		base32[x[9]&0x1F],
		base32[x[10]>>3],
		base32[(x[11]>>6)&0x1F|(x[10]<<2)&0x1F],
		base32[(x[11]>>1)&0x1F],
		base32[(x[11]<<4)&0x1F],
	)
	return dst
}

// String returns a base32 hex lowercased representation of the id.
func (x XID) String() string {
	return b2s(x.Encode(nil))
}

const base32r = "\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18" +
	"\x19\x1a\x1b\x1c\x1d\x1e\x1f\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff"

func ParseXID(s string) (x XID) {
	x[11] = base32r[s[17]]<<6 | base32r[s[18]]<<1 | base32r[s[19]]>>4
	x[10] = base32r[s[16]]<<3 | base32r[s[17]]>>2
	x[9] = base32r[s[14]]<<5 | base32r[s[15]]
	x[8] = base32r[s[12]]<<7 | base32r[s[13]]<<2 | base32r[s[14]]>>3
	x[7] = base32r[s[11]]<<4 | base32r[s[12]]>>1
	x[6] = base32r[s[9]]<<6 | base32r[s[10]]<<1 | base32r[s[11]]>>4
	x[5] = base32r[s[8]]<<3 | base32r[s[9]]>>2
	x[4] = base32r[s[6]]<<5 | base32r[s[7]]
	x[3] = base32r[s[4]]<<7 | base32r[s[5]]<<2 | base32r[s[6]]>>3
	x[2] = base32r[s[3]]<<4 | base32r[s[4]]>>1
	x[1] = base32r[s[1]]<<6 | base32r[s[2]]<<1 | base32r[s[3]]>>4
	x[0] = base32r[s[0]]<<3 | base32r[s[1]]>>2
	return
}
