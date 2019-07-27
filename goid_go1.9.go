// +build arm amd64 amd64p32
// +build go1.9

package log

import (
	"unsafe"
)

func getg() uintptr

func goid() int64 {
	return *(*int64)(unsafe.Pointer(getg() + 152))
}
