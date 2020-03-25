// +build linux

package log

import (
	_ "unsafe"
)

//go:noescape
//go:linkname walltime runtime.walltime
func walltime() (sec int64, nsec int32)
