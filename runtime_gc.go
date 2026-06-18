//go:build gc

package log

import (
	"time"
	_ "unsafe"
)

//go:noescape
//go:linkname absDate time.absDate
func absDate(abs uint64, full bool) (year int, month time.Month, day int, yday int)

//go:noescape
//go:linkname absClock time.absClock
func absClock(abs uint64) (hour, min, sec int)

//go:noescape
//go:linkname caller1 runtime.callers
func caller1(skip int, pc *uintptr, len, cap int) int
