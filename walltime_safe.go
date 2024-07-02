//go:build !((linux && amd64) || (linux && arm64))

package log

import (
	_ "unsafe"
)

//go:noescape
//go:linkname time_now time.now
func time_now() (sec int64, nsec int32, mono int64)

func walltime() (sec int64, nsec int32) {
	sec, nsec, _ = time_now()
	return
}
