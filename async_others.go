//go:build !linux
// +build !linux

package log

func (w *AsyncWriter) vwriter() {
	panic("not_implemented")
}
