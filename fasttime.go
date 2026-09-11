//go:build gc && linux && (amd64 || arm64) && go1.25 && !go1.28

package log

import _ "unsafe"

// The g/m offsets vdsoCallG0 hardcodes come from the runtime of Go 1.25 through
// 1.27: m.curg and m.gsignal moved from 192/80 to 184/72 in Go 1.25. Every
// other toolchain, and every Go release outside that window, takes the
// fasttime_zzz.go fallback instead, because a wrong offset can sit in an
// integer field, and dereferencing it to check it would fault during package
// init instead of disabling the fast path. Raise the upper bound only after
// checking the offsets against the new release.

// timespec mirrors struct timespec on linux/amd64 and linux/arm64.
type timespec struct {
	sec  int64
	nsec int64
}

// vdsoClockgettimeSym is the kernel's __vdso_clock_gettime entry point. The
// runtime keeps it pullable via linkname on purpose (see badlinkname_linux.go).
//
//go:linkname vdsoClockgettimeSym runtime.vdsoClockgettimeSym
var vdsoClockgettimeSym uintptr

// runtimeGetm returns the current m; g0LayoutOK cross-checks the g.m offset
// against it, so a wrong offset can never be used silently.
//
//go:linkname runtimeGetm runtime.getm
func runtimeGetm() uintptr

// vdsoCallG0 switches to the g0 stack, performs a single vDSO
// CLOCK_REALTIME read and switches back. Implemented per architecture.
// It returns zero on success; on failure it leaves ts unchanged.
//
// A caller that is not running on m.curg (already on g0 or gsignal) fails
// deliberately: the runtime finds g through a stash on the gsignal stack when
// a signal lands inside the vDSO, and this path only writes that stash when
// switching stacks, so such callers are routed through now() instead, whose
// runtime implementation does its own stack switch and signal bookkeeping.
//
// Unlike the runtime this does not set m.vdsoPC/m.vdsoSP. Doing that would
// need two more hardcoded m offsets, and unlike the g0 offsets they could not
// be validated cheaply: a wrong one would corrupt runtime state instead of
// merely disabling the fast path. The cost is traceback fidelity if
// SIGPROF/SIGQUIT lands inside the handful of instructions spent in the vDSO.
//
//go:noescape
func vdsoCallG0(fn uintptr, ts *timespec) int32

// g0LayoutOK reports whether the g/m field offsets hardcoded in the
// architecture specific assembly still match this toolchain: it walks
// g.m -> m.curg -> m.g0 and checks that the saved SP of g0 lies inside the
// stack of g0. m must be the value returned by runtimeGetm, and any wrong m is
// rejected without dereferencing it. Implemented per architecture.
//
// m.gsignal, which vdsoCallG0 also reads, is deliberately not checked: on a
// mismatched layout that slot can hold a non-pointer, and dereferencing it to
// tell the two apart would fault during package init instead of disabling the
// fast path.
//
//go:noescape
func g0LayoutOK(m uintptr) bool

var vdsoReady = vdsoClockgettimeSym != 0 && g0LayoutOK(runtimeGetm())

// walltime returns sec and nsec of CLOCK_REALTIME with the same precision
// as time.Now, but pays for a single clock read. time.now always reads both
// CLOCK_REALTIME (walltime) and CLOCK_MONOTONIC (nanotime), and the header only
// ever needs the former.
//
// The read is done on the g0 stack because the kernel's clock_gettime may use
// up to a page of stack (stack probes in gettime_sym on hardened kernels), which
// can hit the guard page of a goroutine stack. It also has to run on g0 for the
// signal handling the runtime normally arranges.
//
// When the vDSO is unavailable, the offsets no longer match this toolchain, or
// the vDSO call fails, this falls back to now() and keeps the old behavior.
func walltime() (sec int64, nsec int32) {
	var ts timespec
	if vdsoReady && vdsoCallG0(vdsoClockgettimeSym, &ts) == 0 {
		return ts.sec, int32(ts.nsec)
	}
	sec, nsec, _ = now()
	return
}
