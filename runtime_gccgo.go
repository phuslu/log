//go:build gccgo

package log

import (
	"strings"
	"sync"
	"time"
	"unsafe"
)

const (
	secondsPerMinute = 60
	secondsPerHour   = 60 * secondsPerMinute
	secondsPerDay    = 24 * secondsPerHour
	daysPer400Years  = 365*400 + 97
	daysPer100Years  = 365*100 + 24
	daysPer4Years    = 365*4 + 1
	absoluteZeroYear = -292277022399
)

var gccgoDaysBefore = [...]int32{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334, 365}

// gccgoG matches libgo's runtime.g through the goid field.
type gccgoG struct {
	_panic       unsafe.Pointer
	_defer       unsafe.Pointer
	m            unsafe.Pointer
	syscallsp    uintptr
	syscallpc    uintptr
	param        unsafe.Pointer
	atomicstatus uint32
	goid         int64
}

//go:noescape
//go:linkname gccgoGetg runtime.getg
func gccgoGetg() *gccgoG

func goid() int {
	return int(gccgoGetg().goid)
}

// Goid returns the current goroutine id.
// It exactly matches goroutine id of the stack trace.
func Goid() int64 {
	return int64(goid())
}

// BSD-3-Clause, copied from Go's time.absDate.
func absDate(abs uint64, full bool) (year int, month time.Month, day int, yday int) {
	d := abs / secondsPerDay
	n := d / daysPer400Years
	y := 400 * n
	d -= daysPer400Years * n
	n = d / daysPer100Years
	n -= n >> 2
	y += 100 * n
	d -= daysPer100Years * n
	n = d / daysPer4Years
	y += 4 * n
	d -= daysPer4Years * n
	n = d / 365
	n -= n >> 2
	y += n
	d -= 365 * n
	year = int(int64(y) + absoluteZeroYear)
	yday = int(d)
	if !full {
		return
	}

	day = yday
	if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
		switch {
		case day > 31+29-1:
			day--
		case day == 31+29-1:
			return year, time.February, 29, yday
		}
	}

	month = time.Month(day / 31)
	begin := int(gccgoDaysBefore[month])
	if end := int(gccgoDaysBefore[month+1]); day >= end {
		month++
		begin = end
	}
	month++
	day = day - begin + 1
	return
}

func absClock(abs uint64) (hour, min, sec int) {
	sec = int(abs % secondsPerDay)
	hour = sec / secondsPerHour
	sec -= hour * secondsPerHour
	min = sec / secondsPerMinute
	sec -= min * secondsPerMinute
	return
}

// gccgoLocation must match runtime.location.
type gccgoLocation struct {
	pc       uintptr
	filename string
	function string
	lineno   int
}

//go:noescape
//extern runtime_callers
func gccgoCallers(skip int32, locbuf *gccgoLocation, max int32, keepThunks bool) int32

//go:noescape
//go:linkname gccgoFuncFileLine runtime.funcfileline
func gccgoFuncFileLine(pc uintptr, index int32, more bool) (name, file string, line, frames int)

//go:noescape
//go:linkname gccgoDecodeIdentifier runtime.decodeIdentifier
func gccgoDecodeIdentifier([]byte) int

//go:noinline
func caller1(skip int, pc *uintptr, length, capacity int) int {
	if pc == nil || skip < 0 || length < 1 || capacity < 1 {
		return 0
	}
	var location gccgoLocation
	if gccgoCallers(int32(skip+1), &location, 1, false) != 1 {
		return 0
	}
	*pc = location.pc
	return 1
}

// Fastrandn returns a pseudorandom uint32 in [0,n).
//
//go:noescape
//go:linkname Fastrandn runtime.fastrandn
func Fastrandn(n uint32) uint32

func pcFileLine(pc uintptr) (file string, line int) {
	if pc == 0 {
		return "", 0
	}
	_, file, line, _ = gccgoFuncFileLine(pc-1, -1, false)
	return file, line
}

func pcFileLineName(pc uintptr) (file string, line int, name string) {
	if pc == 0 {
		return "", 0, ""
	}
	name, file, line, _ = gccgoFuncFileLine(pc-1, -1, false)
	return file, line, demangleGCCGoSymbol(name)
}

var gccgoSymbolCache sync.Map

// demangleGCCGoSymbol reverses gccgo's underscore encoding. Libgo's low-level
// symbol lookup returns the encoded name; CallersFrames normally decodes it.
func demangleGCCGoSymbol(name string) string {
	if !strings.Contains(name, ".") || !strings.Contains(name, "_") {
		return name
	}
	if decoded, ok := gccgoSymbolCache.Load(name); ok {
		return decoded.(string)
	}

	b := []byte(name)
	decoded := string(b[:gccgoDecodeIdentifier(b)])
	actual, _ := gccgoSymbolCache.LoadOrStore(name, decoded)
	return actual.(string)
}
