//go:build gccgo

package log

import (
	"strings"
	"sync"
	"unsafe"
)

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

//go:noescape
//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64)

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
