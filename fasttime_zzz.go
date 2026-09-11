//go:build !(gc && linux && (amd64 || arm64) && go1.25 && !go1.28)

package log

// walltime has no vDSO fast path on this platform, with this compiler (gccgo
// has no Go assembler and no runtime vdsoClockgettimeSym), or outside the Go
// releases fasttime.go is written for, so it reads the same clock
// time.now does.
func walltime() (sec int64, nsec int32) {
	sec, nsec, _ = now()
	return
}
