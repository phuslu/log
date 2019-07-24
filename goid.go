// +build !go1.9,!amd64,!amd64p32,!arm

package log

func goid() int64 {
	return 0
}
