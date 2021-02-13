package log

import (
	"time"
)

func timewall(t time.Time) (sec int64, nsec int32) {
	sec = t.Unix()
	nsec = int32(t.Nanosecond())
	return
}
