// +build !linux

package log

import (
	"time"
)

func walltime() (sec int64, nsec int32) {
	now := time.Now()
	sec = now.Unix()
	nsec = int32(now.Nanosecond())
	return
}
