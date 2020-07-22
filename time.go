// +build !linux

package log

import (
	"time"
)

func TimestampMS() int64 {
	now := time.Now()
	return now.Unix()*1000 + int64(now.Nanosecond())/1000000
}

func walltime() (sec int64, nsec int32) {
	now := time.Now()
	sec = now.Unix()
	nsec = int32(now.Nanosecond())
	return
}
