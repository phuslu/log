// +build !linux

package log

import (
	"time"
)

func walltime() (sec int64, nsec int32) {
	sec, nsec = timewall(time.Now())
	return
}
