// +build amd64 arm64

package log

func goid() int64

// Goid returns the current goroutine id.
// It exactly matches goroutine id of the stack trace.
func Goid() int64 {
	return goid()
}
