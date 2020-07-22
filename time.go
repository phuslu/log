// +build !linux

package log

import (
	"time"
)

func TimestampMS() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func walltime() (sec int64, nsec int32) {
	now := time.Now()
	sec = now.Unix()
	nsec = int32(now.Nanosecond())
	return
}
