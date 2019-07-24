// +build arm amd64 amd64p32
// +build go1.9

package log

type g struct {
	stack struct {
		lo uintptr
		hi uintptr
	}
	stackguard0 uintptr
	stackguard1 uintptr
	_panic      uintptr
	_defer      uintptr
	m           uintptr
	sched       struct {
		sp   uintptr
		pc   uintptr
		g    uintptr
		ctxt uintptr
		ret  uintptr
		lr   uintptr
		bp   uintptr
	}
	syscallsp    uintptr
	syscallpc    uintptr
	stktopsp     uintptr
	param        uintptr
	atomicstatus uint32
	stackLock    uint32
	goid         int64
}

func getg() *g

func goid() int64 {
	return getg().goid
}
