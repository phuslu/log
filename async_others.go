//go:build !linux
// +build !linux

package log

func (w *AsyncWriter) writever() {
	panic("not_implemented")
}
