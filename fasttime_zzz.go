//go:build !(gc && linux && (amd64 || arm64) && go1.25 && !go1.28)

package log

// walltime has no vDSO fast path on this platform, with this compiler (gccgo
// has no Go assembler and no runtime vdsoClockgettimeSym), or outside the Go
// releases fasttime.go is written for, so it returns zero and callers fall back
// to now().
func walltime() (sec int64, nsec int32) {
	return 0, 0
}
