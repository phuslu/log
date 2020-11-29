// +build amd64 arm64 arm

package log

// Goid returns the current goroutine id.
// It exactly matches goroutine id of the stack trace.
func Goid() int64 {
	return getg().goid
}

func getg() *g
