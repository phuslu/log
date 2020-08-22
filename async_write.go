// +build !linux
// +build !darwin
// +build !freebsd
// +build !openbsd
// +build !netbsd
// +build !dragonfly

package log

// writev panic io.Writer.
func (w *AsyncWriter) writev(p []byte) (n int, err error) {
	panic("not implemented")
}
