//go:build go1.18 && !go1.22
// +build go1.18,!go1.22

package log

import (
	_ "unsafe"
)

// Fastrandn returns a pseudorandom uint32 in [0,n).
//
//go:noescape
//go:linkname Fastrandn runtime.fastrandn
func Fastrandn(x uint32) uint32
