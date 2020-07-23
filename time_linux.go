// +build linux

package log

import (
	_ "unsafe" // for runtime.walltime
)

// TimestampMS returns Unix timestamp integers in microseconds.
func TimestampMS() int64 {
	sec, nsec := walltime()
	return sec*1000 + int64(nsec)/1000000
}

//go:noescape
//go:linkname walltime runtime.walltime
func walltime() (sec int64, nsec int32)
