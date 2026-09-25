//go:build gc

package log

import (
	_ "unsafe"
)

// Fastrandn returns a pseudorandom uint32 in [0,n).
//
//go:noescape
//go:linkname Fastrandn runtime.cheaprandn
func Fastrandn(n uint32) uint32

//go:noescape
//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64)

//go:noescape
//go:linkname caller1 runtime.callers
func caller1(skip int, pc *uintptr, len, cap int) int
