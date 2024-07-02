//go:build (linux && amd64) || (linux && arm64)

package log

func walltime() (sec int64, nsec int32)
