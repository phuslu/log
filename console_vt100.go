// +build !windows

package log

func (w *ConsoleWriter) Write(p []byte) (int, error) {
	return w.vtWrite(p)
}
