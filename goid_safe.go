// +build !amd64,!amd64p32,!arm,!arm64
// +build !go1.9

package log

import (
	"runtime"
)

func goid() (n int64) {
	const offset = len("goroutine ")
	var data [32]byte
	b := data[:runtime.Stack(data[:], false)]
	if len(b) <= offset {
		return
	}
	for i := offset; i < len(b); i++ {
		j := int64(b[i] - '0')
		if j < 0 || j > 9 {
			break
		}
		n = n*10 + j
	}
	return n
}
