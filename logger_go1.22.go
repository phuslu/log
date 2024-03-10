//go:build go1.22
// +build go1.22

package log

import (
	_ "unsafe"
)

// Fastrandn returns a pseudorandom uint32 in [0,n).
//
//go:noescape
//go:linkname Fastrandn runtime.cheaprandn
func Fastrandn(x uint32) uint32
