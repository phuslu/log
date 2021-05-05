package log

import (
	"time"
	_ "unsafe" // for time.now
)

//go:noescape
//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64)

func timewall(t time.Time) (sec int64, nsec int32) {
	sec = t.Unix()
	nsec = int32(t.Nanosecond())
	return
}
