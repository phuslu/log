// +build !go1.9,!amd64,!amd64p32,!arm

package log

import (
	"runtime"
)

func goid() (n int64) {
	const offset = len("goroutine ")
	var data [64]byte
	b := data[:runtime.Stack(data[:], false)]
	if len(b) <= offset {
		return
	}
	for i := offset; b[i] != ' '; i++ {
		n = n*10 + int64(b[i]-'0')
	}
	return n
}
